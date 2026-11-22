// The Manage project screen (admins): one catalog app run from here through
// the portal's admin key there. Its own accounts (make one, a guest, a login
// link, role, expiry, disable, sign out everywhere, delete), its invite links,
// and API keys for its accounts. What an app supports comes from its last
// check (portal contract v2, CROSS_PROJECT.md); what it lacks is explained,
// not offered. app.js calls Manage.init( me ) for admins.
const Manage = {
  projectId: null,
  projects: [],
  // The rows the people table shows (narrowed by its search), and every
  // account on the project, for the API keys account picker.
  people: [],
  accounts: [],
  // The project the tables now hold, so a switch empties them at once
  // instead of leaving the last project's rows under the new name.
  shownId: null,

  init( me ) {
    if ( me.role !== "admin" ) return;
    Dom.show( Dom.get( "manage-panel" ) , true );
    Dom.get( "manage-project" ).addEventListener( "change" , ( event ) => {
      this.projectId = Number( event.target.value ) || null;
      this.load();
    } );
    Dom.typed( Dom.get( "manage-search" ) , () => this.refreshPeople() );
    Dom.onSubmit( Dom.get( "manage-new-form" ) , this.onNew.bind( this ) );
    Dom.get( "manage-guest-button" ).addEventListener( "click" , () => this.guestPreset() );
    Dom.onSubmit( Dom.get( "manage-invite-form" ) , this.onInvite.bind( this ) );
    Dom.onSubmit( Dom.get( "manage-key-form" ) , this.onMint.bind( this ) );
    Dom.get( "manage-key-account" ).addEventListener( "change" , () => this.refreshKeys() );
    Dom.get( "manage-tab-invites" ).addEventListener( "shown.bs.tab" , () => this.refreshInvites() );
    Dom.get( "manage-tab-keys" ).addEventListener( "shown.bs.tab" , () => this.refreshKeys() );
    window.addEventListener( "nav" , () => {
      if ( Nav.current() === "manage" ) this.fillProjects();
    } );
    // Nothing to load yet: Nav.init's first "nav" does it if this screen opens.
  },

  // The project picker, keeping the one picked while it still exists.
  async fillProjects() {
    let page;
    try {
      page = await Api.listCatalog( 0 , { limit: 1000 } );
    } catch ( error ) {
      App.showError( error.message );
      return;
    }
    this.projects = page.projects;
    const select = Dom.get( "manage-project" );
    Dom.swap( select , ...page.projects.map( ( project ) => Dom.el( "option" , { text: project.name , attrs: { value: String( project.id ) } } ) ) );
    const ids = page.projects.map( ( project ) => project.id );
    if ( !ids.includes( this.projectId ) ) this.projectId = ids.length ? ids[ 0 ] : null;
    if ( this.projectId ) select.value = String( this.projectId );
    Dom.show( Dom.get( "manage-body" ) , this.projectId !== null );
    Dom.show( Dom.get( "manage-none" ) , this.projectId === null );
    if ( this.projectId ) await this.load();
  },

  // From a Catalog row's Manage button.
  // Nav.show sends "nav", whose listener fills the picker and loads it.
  showFor( projectId ) {
    this.projectId = projectId;
    Nav.show( "manage" );
  },

  project() {
    return this.projects.find( ( project ) => project.id === this.projectId ) || { capabilities: [] };
  },

  has( capability ) {
    return ( this.project().capabilities || [] ).includes( capability );
  },

  // Roles this app takes; viewer only where it has view-only accounts.
  roles( withAdmin ) {
    const roles = [ "user" ];
    if ( this.has( "viewer_role" ) ) roles.push( "viewer" );
    if ( withAdmin ) roles.push( "admin" );
    return roles;
  },

  fillRoles( select , withAdmin , picked ) {
    const roles = this.roles( withAdmin );
    Dom.swap( select , ...roles.map( ( role ) => Dom.el( "option" , { text: I18n.get( "manage.role_" + role ) , attrs: { value: role } } ) ) );
    select.value = roles.includes( picked ) ? picked : "user";
  },

  async load() {
    if ( !this.projectId ) return;
    if ( this.shownId !== this.projectId ) {
      this.shownId = this.projectId;
      [ "manage-people-table" , "manage-invites-table" , "manage-keys-table" ].forEach( ( table ) => Pager.reset( table ) );
      this.people = [];
      this.accounts = [];
      this.fillKeyAccounts();
    }
    this.showSupport();
    this.fillRoles( Dom.get( "manage-new-role" ) , true );
    this.fillRoles( Dom.get( "manage-invite-role" ) , false );
    this.fillRoles( Dom.get( "manage-key-role" ) , true );
    Dom.get( "manage-new-days" ).disabled = !this.has( "account_expiry" );
    Dom.get( "manage-guest-button" ).disabled = !this.has( "viewer_role" ) || !this.has( "account_expiry" );
    await this.refreshAll();
    if ( Dom.get( "manage-tab-invites" ).classList.contains( "active" ) ) await this.refreshInvites();
    if ( Dom.get( "manage-tab-keys" ).classList.contains( "active" ) ) await this.refreshKeys();
  },

  // What the app can do from here, and what an upgrade would add.
  showSupport() {
    const project = this.project();
    const all = [ "viewer_role" , "account_expiry" , "invite_links" , "key_minting" , "key_sign_in" ];
    Dom.swap( Dom.get( "manage-capabilities" ) , ...all.map( ( capability ) => {
      const on = this.has( capability );
      const badge = Dom.badge( I18n.get( "manage.cap_" + capability ) , on ? "success" : "secondary" );
      badge.prepend( Dom.el( "i" , { class: "bi " + ( on ? "bi-check-lg" : "bi-dash" ) + " me-1" , attrs: { "aria-hidden": "true" } } ) );
      badge.title = I18n.get( on ? "manage.cap_on" : "manage.cap_off" );
      return badge;
    } ) );
    const missing = all.filter( ( capability ) => !this.has( capability ) );
    const behind = Dom.get( "manage-behind" );
    const reason = project.key_status && project.key_status !== "ok" ? I18n.get( "manage.behind_key" )
      : missing.length ? I18n.get( "manage.behind" ).replace( "{version}" , () => project.version || "?" ) : "";
    Dom.text( behind , reason );
    Dom.show( behind , reason !== "" );
    this.gate( "invites" , "invite_links" );
    this.gate( "keys" , "key_minting" , this.mintingBlocked() ? "manage.keys_no_grant" : null );
  },

  // An app that says what the portal's key may do, and it may not make keys.
  mintingBlocked() {
    const key = this.project().key || {};
    return this.has( "key_grants" ) && key.grants && !key.may_mint_keys;
  },

  // A pane shows only where the app has capability; blockedKey, when given,
  // keeps it shut anyway and says why.
  gate( pane , capability , blockedKey ) {
    const on = this.has( capability ) && !blockedKey;
    Dom.show( Dom.get( "manage-" + pane + "-on" ) , on );
    const off = Dom.get( "manage-" + pane + "-off" );
    Dom.text( off , I18n.get( blockedKey || "manage." + pane + "_off" ).replace( /\{name\}/g , () => this.project().name || "" ) );
    Dom.show( off , !on );
  },

  // --- people ------------------------------------------------------------------

  async refreshPeople() {
    if ( !this.projectId ) return;
    const projectId = this.projectId;
    const q = Dom.get( "manage-search" ).value.trim();
    const body = Dom.get( "manage-people-table" ).querySelector( "tbody" );
    this.people = [];
    await Pager.load( "manage-people-table" , "users" , ( offset ) => Api.remoteUsers( projectId , q , offset ) , ( user ) => {
      this.people.push( user );
      body.appendChild( this.personRow( user ) );
    } );
  },

  // The table and the key account picker, after something changed accounts.
  async refreshAll() {
    await Promise.all( [ this.refreshPeople() , this.refreshAccounts() ] );
  },

  // Every account on the project, all pages, whatever the search box says:
  // the key account picker lists them all. Errors show on the table's load.
  async refreshAccounts() {
    const projectId = this.projectId;
    if ( !projectId ) return;
    const accounts = [];
    let offset = 0;
    try {
      for ( let pages = 0 ; pages < 50 ; pages++ ) {
        const page = await Api.remoteUsers( projectId , "" , offset );
        accounts.push( ...page.users );
        if ( page.next === null || page.next === undefined ) break;
        offset = page.next;
      }
    } catch ( error ) {
      return;
    }
    if ( projectId !== this.projectId ) return;
    this.accounts = accounts;
    this.fillKeyAccounts();
  },

  personRow( user ) {
    const name = Dom.el( "td" , { children: [
      Dom.el( "div" , { class: "fw-semibold text-break" , text: user.display_name } ),
      user.portal_name ? Dom.el( "div" , { class: "small text-body-secondary" ,
        text: I18n.get( "manage.portal_person" ).replace( "{name}" , () => user.portal_name ) } ) : null,
    ].filter( Boolean ) } );
    return Dom.el( "tr" , { children: [
      name,
      Dom.el( "td" , { children: [ Dom.badge( I18n.get( "manage.role_" + user.role ) || user.role , user.role === "admin" ? "primary" : user.role === "viewer" ? "info" : "secondary" ) ] } ),
      Dom.el( "td" , { children: this.personStatus( user ) } ),
      Dom.el( "td" , { children: [ this.personActions( user ) ] } ),
    ] } );
  },

  personStatus( user ) {
    const parts = [];
    if ( user.owner ) parts.push( Dom.badge( I18n.get( "manage.owner" ) , "dark" ) );
    if ( user.portal ) parts.push( Dom.badge( I18n.get( "manage.this_portal" ) , "primary" ) );
    else if ( user.kind === "service" ) parts.push( Dom.badge( I18n.get( "manage.service" ) , "info" ) );
    if ( user.expired ) parts.push( Dom.badge( I18n.get( "manage.expired" ) , "danger" ) );
    else if ( user.disabled ) parts.push( Dom.badge( I18n.get( "manage.disabled" ) , "warning" ) );
    else parts.push( Dom.badge( I18n.get( "manage.active" ) , "success" ) );
    if ( user.expires_at && !user.expired ) {
      parts.push( Dom.el( "div" , { class: "small text-body-secondary" ,
        text: I18n.get( "manage.expires_on" ).replace( "{date}" , () => new Date( user.expires_at ).toLocaleDateString() ) } ) );
    }
    return parts;
  },

  personActions( user ) {
    const item = ( key , handler , danger ) => Dom.el( "button" , {
      class: "btn btn-sm " + ( danger ? "btn-outline-danger" : "btn-outline-secondary" ),
      attrs: { type: "button" },
      text: I18n.get( key ),
      on: { click: () => handler( user ) },
    } );
    if ( user.owner ) return Dom.el( "span" , { class: "small text-body-secondary" , text: I18n.get( "manage.owner_note" ) } );
    if ( user.acts === false ) return Dom.el( "span" , { class: "small text-body-secondary" , text: I18n.get( "manage.admin_note" ) } );
    return Dom.rowActions( user.display_name , [
      item( "manage.link_button" , ( row ) => this.onLoginLink( row ) ),
      item( "manage.role_button" , ( row ) => this.onRole( row ) ),
      this.has( "account_expiry" ) ? item( "manage.expiry_button" , ( row ) => this.onExpiry( row ) ) : null,
      this.has( "key_minting" ) && !this.mintingBlocked() ? item( "manage.keys_button" , ( row ) => this.showKeysFor( row ) ) : null,
      item( user.disabled ? "manage.enable_button" : "manage.disable_button" , ( row ) => this.onDisabled( row ) ),
      item( "manage.revoke_button" , ( row ) => this.onRevoke( row ) , true ),
      item( "manage.delete_button" , ( row ) => this.onDelete( row ) , true ),
    ].filter( Boolean ) , 1 );
  },

  // The fields of the new-account form as a guest: view only, a week.
  guestPreset() {
    Dom.get( "manage-new-role" ).value = "viewer";
    Dom.get( "manage-new-days" ).value = "7";
    const name = Dom.get( "manage-new-name" );
    if ( !name.value.trim() ) name.value = I18n.get( "manage.guest_name" );
    name.focus();
    name.select();
  },

  async onNew( event ) {
    event.preventDefault();
    const name = Dom.get( "manage-new-name" ).value.trim();
    if ( !name ) {
      Dom.fieldError( Dom.get( "manage-new-name" ) , I18n.get( "manage.name_needed" ) );
      return;
    }
    const days = Number( Dom.get( "manage-new-days" ).value ) || 0;
    const project = this.project();
    try {
      const answer = await Api.createRemoteUser( this.projectId , { display_name: name , role: Dom.get( "manage-new-role" ).value , expires_in_days: days } );
      Dom.get( "manage-new-form" ).reset();
      Dom.get( "manage-new-role" ).value = "user";
      await this.handLink( project , name , answer.login_url );
      await this.refreshAll();
    } catch ( error ) { App.showError( error.message ); }
  },

  handLink( project , name , url ) {
    return Dom.handOver( {
      value: url,
      copied: I18n.get( "manage.link_copied" ).replace( "{name}" , () => name ),
      title: I18n.get( "manage.link_heading" ).replace( "{name}" , () => name ).replace( "{project}" , () => project.name ),
      label: I18n.get( "projects.new_link_label" ),
      help: I18n.get( "projects.new_link_help" ),
    } );
  },

  async onLoginLink( user ) {
    try {
      const answer = await Api.remoteLoginLink( this.projectId , user.id );
      await this.handLink( this.project() , user.display_name , answer.login_url );
    } catch ( error ) { App.showError( error.message ); }
  },

  async onRole( user ) {
    const roles = this.roles( true );
    const saved = await Dom.form( {
      title: I18n.get( "manage.role_heading" ).replace( "{name}" , () => user.display_name ),
      fields: [ { name: "role" , type: "select" , label: I18n.get( "manage.role_label" ) , help: I18n.get( "manage.role_help" ) ,
        options: roles.map( ( role ) => ( { value: role , label: I18n.get( "manage.role_" + role ) } ) ) , value: user.role } ],
      save: ( values ) => Api.setRemoteRole( this.projectId , user.id , values.role ),
    } );
    if ( saved !== null ) await this.refreshAll();
  },

  async onExpiry( user ) {
    const saved = await Dom.form( {
      title: I18n.get( "manage.expiry_heading" ).replace( "{name}" , () => user.display_name ),
      fields: [ { name: "days" , type: "number" , label: I18n.get( "manage.days_label" ) , help: I18n.get( "manage.expiry_help" ) , min: 0 , max: 3650 , value: "" } ],
      save: ( values ) => Api.setRemoteExpiry( this.projectId , user.id , values.days || 0 ),
    } );
    if ( saved !== null ) await this.refreshPeople();
  },

  async onDisabled( user ) {
    const disabling = !user.disabled;
    if ( disabling && !( await Dom.confirm( I18n.get( "manage.disable_confirm" ).replace( "{name}" , () => user.display_name ) ) ) ) return;
    try {
      await Api.setRemoteDisabled( this.projectId , user.id , disabling );
      await this.refreshPeople();
    } catch ( error ) { App.showError( error.message ); }
  },

  async onRevoke( user ) {
    if ( !( await Dom.confirm( I18n.get( "manage.revoke_confirm" ).replace( "{name}" , () => user.display_name ) ) ) ) return;
    try {
      await Api.revokeRemoteAccess( this.projectId , user.id );
      Dom.toast( I18n.get( "manage.revoked" ).replace( "{name}" , () => user.display_name ) , "ok" );
    } catch ( error ) { App.showError( error.message ); }
  },

  async onDelete( user ) {
    const question = I18n.get( "manage.delete_confirm" ).replace( "{name}" , () => user.display_name ).replace( "{project}" , () => this.project().name );
    if ( !( await Dom.confirm( question ) ) ) return;
    try {
      await Api.deleteRemoteUser( this.projectId , user.id );
      await this.refreshAll();
    } catch ( error ) { App.showError( error.message ); }
  },

  // --- invite links ----------------------------------------------------------------

  async refreshInvites() {
    if ( !this.projectId || !this.has( "invite_links" ) ) return;
    const projectId = this.projectId;
    const body = Dom.get( "manage-invites-table" ).querySelector( "tbody" );
    await Pager.load( "manage-invites-table" , "invite_links" , () => Api.remoteInviteLinks( projectId ) , ( link ) => {
      const status = link.revoked ? Dom.badge( I18n.get( "manage.revoked_badge" ) , "secondary" )
        : link.live ? null : Dom.badge( I18n.get( "manage.expired" ) , "secondary" );
      const revoke = link.live ? Dom.rowActions( link.label , [ Dom.el( "button" , {
        class: "btn btn-sm btn-outline-danger" , attrs: { type: "button" } , text: I18n.get( "manage.revoke_link_button" ) ,
        on: { click: () => this.onRevokeInvite( link ) } } ) ] ) : null;
      body.appendChild( Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ Dom.el( "div" , { class: "fw-semibold text-break" , text: link.label } ) , status ].filter( Boolean ) } ),
        Dom.el( "td" , { text: link.used + " / " + link.max_uses } ),
        Dom.el( "td" , { text: I18n.get( "manage.role_" + ( link.role || "user" ) ) } ),
        Dom.el( "td" , { children: [ Dom.when( link.expires_at , "" ) ] } ),
        Dom.el( "td" , { children: [ revoke ].filter( Boolean ) } ),
      ] } ) );
    } );
  },

  async onInvite( event ) {
    event.preventDefault();
    const label = Dom.get( "manage-invite-label" ).value.trim();
    if ( !label ) {
      Dom.fieldError( Dom.get( "manage-invite-label" ) , I18n.get( "manage.label_needed" ) );
      return;
    }
    try {
      const answer = await Api.createRemoteInvite( this.projectId , {
        label: label,
        max_uses: Number( Dom.get( "manage-invite-uses" ).value ) || 1,
        days: Number( Dom.get( "manage-invite-days" ).value ) || 1,
        role: Dom.get( "manage-invite-role" ).value,
      } );
      Dom.get( "manage-invite-label" ).value = "";
      await Dom.handOver( {
        show: true , value: answer.join_url,
        title: I18n.get( "manage.invite_made" ).replace( "{name}" , () => label ),
        label: I18n.get( "manage.invite_link_label" ),
        help: I18n.get( "manage.invite_link_help" ),
      } );
      await this.refreshInvites();
    } catch ( error ) { App.showError( error.message ); }
  },

  async onRevokeInvite( link ) {
    if ( !( await Dom.confirm( I18n.get( "manage.revoke_link_confirm" ).replace( "{name}" , () => link.label ) ) ) ) return;
    try {
      await Api.revokeRemoteInvite( this.projectId , link.id );
      await this.refreshInvites();
    } catch ( error ) { App.showError( error.message ); }
  },

  // --- API keys ---------------------------------------------------------------------

  // The accounts a key can be made for: everyone listed but the owner.
  fillKeyAccounts() {
    const select = Dom.get( "manage-key-account" );
    const picked = select.value;
    // Not the owner, not the portal's own account, never a service account,
    // and not an admin the portal's key may not run.
    const accounts = this.accounts.filter( ( user ) => !user.owner && !user.portal && user.kind !== "service" && user.acts !== false );
    Dom.swap( select , ...accounts.map( ( user ) => Dom.el( "option" , { text: user.display_name , attrs: { value: String( user.id ) } } ) ) );
    if ( accounts.some( ( user ) => String( user.id ) === picked ) ) select.value = picked;
    if ( select.value !== picked ) this.refreshKeys();
  },

  async showKeysFor( user ) {
    Dom.get( "manage-key-account" ).value = String( user.id );
    bootstrap.Tab.getOrCreateInstance( Dom.get( "manage-tab-keys" ) ).show();
    await this.refreshKeys();
  },

  keyAccount() {
    const id = Number( Dom.get( "manage-key-account" ).value ) || null;
    return this.accounts.find( ( user ) => user.id === id ) || null;
  },

  async refreshKeys() {
    if ( !this.projectId || !this.has( "key_minting" ) ) return;
    const account = this.keyAccount();
    Dom.text( Dom.get( "manage-keys-heading" ) , account
      ? I18n.get( "manage.keys_heading" ).replace( "{name}" , () => account.display_name ) : I18n.get( "manage.keys_heading_none" ) );
    const body = Dom.get( "manage-keys-table" ).querySelector( "tbody" );
    if ( !account ) {
      Dom.clear( body );
      return;
    }
    const projectId = this.projectId;
    await Pager.load( "manage-keys-table" , "keys" , () => Api.remoteKeys( projectId , account.id ) , ( key ) => {
      const status = key.live ? Dom.badge( I18n.get( "manage.active" ) , "success" )
        : Dom.badge( I18n.get( key.revoked ? "manage.revoked_badge" : "manage.expired" ) , "secondary" );
      const minted = key.minted_by_key ? Dom.el( "div" , { class: "small text-body-secondary" , text: I18n.get( "manage.key_from_portal" ) } ) : null;
      body.appendChild( Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ Dom.el( "div" , { class: "fw-semibold text-break" , text: key.name } ) , minted ].filter( Boolean ) } ),
        Dom.el( "td" , { text: I18n.get( "manage.role_" + key.role ) || key.role } ),
        Dom.el( "td" , { children: [ status ] } ),
        Dom.el( "td" , { children: [ Dom.when( key.expires_at , I18n.get( "manage.never" ) ) ] } ),
        Dom.el( "td" , { children: key.live ? [ Dom.rowActions( key.name , [ Dom.el( "button" , {
          class: "btn btn-sm btn-outline-danger" , attrs: { type: "button" } , text: I18n.get( "manage.revoke_key_button" ) ,
          on: { click: () => this.onRevokeKey( key ) } } ) ] ) ] : [] } ),
      ] } ) );
    } );
  },

  async onMint( event ) {
    event.preventDefault();
    const account = this.keyAccount();
    const name = Dom.get( "manage-key-name" ).value.trim();
    if ( !account ) return;
    if ( !name ) {
      Dom.fieldError( Dom.get( "manage-key-name" ) , I18n.get( "manage.name_needed" ) );
      return;
    }
    try {
      const minted = await Api.mintRemoteKey( this.projectId , account.id , {
        name: name , role: Dom.get( "manage-key-role" ).value , expires_in_days: Number( Dom.get( "manage-key-days" ).value ) || 90,
      } );
      Dom.get( "manage-key-name" ).value = "";
      await Dom.handOver( {
        show: true , value: minted.credential,
        title: I18n.get( "manage.key_made" ).replace( "{name}" , () => account.display_name ),
        label: I18n.get( "manage.key_label" ),
        help: I18n.get( "manage.key_help" ),
      } );
      await this.refreshKeys();
    } catch ( error ) { App.showError( error.message ); }
  },

  async onRevokeKey( key ) {
    if ( !( await Dom.confirm( I18n.get( "manage.revoke_key_confirm" ).replace( "{name}" , () => key.name ) ) ) ) return;
    try {
      await Api.revokeRemoteKey( this.projectId , key.id );
      await this.refreshKeys();
    } catch ( error ) { App.showError( error.message ); }
  },
};
