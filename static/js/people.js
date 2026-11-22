// The People screen (admins): one real person and their accounts in every
// app, this portal included (project 0). Everything comes from one GET of
// /api/admin/people, filled from the people-sync job's last look, so the
// search here filters what is already loaded.
const People = {
  data: null,
  // After Refresh now, look again until the sync has had time to finish.
  pollMs: 4000,
  pollsLeft: 0,
  timer: null,

  async init( me ) {
    if ( me.role !== "admin" ) return;
    Dom.show( Dom.get( "people-panel" ) , true );
    Dom.get( "people-sync" ).addEventListener( "click" , () => this.onSync() );
    Dom.typed( Dom.get( "people-search" ) , () => this.render() );
    window.addEventListener( "nav" , () => {
      if ( Nav.current() === "people" ) this.refresh();
    } );
    if ( Nav.current() === "people" ) await this.refresh();
  },

  async refresh() {
    try {
      this.data = await Api.people();
    } catch ( error ) {
      App.showError( error.message );
      return;
    }
    this.render();
  },

  async onSync() {
    try {
      await Api.syncPeople();
      Dom.toast( I18n.get( "people.sync_started" ) , "ok" );
    } catch ( error ) {
      App.showError( error.message );
      return;
    }
    this.pollsLeft = 8;
    this.poll();
  },

  poll() {
    window.clearTimeout( this.timer );
    if ( this.pollsLeft <= 0 ) return;
    this.pollsLeft -= 1;
    this.timer = window.setTimeout( async () => {
      await this.refresh();
      if ( Nav.current() === "people" ) this.poll();
    } , this.pollMs );
  },

  // --- pieces ----------------------------------------------------------------

  ref( account ) {
    return { project_id: account.project_id , remote_user_id: account.remote_user_id };
  },

  words( key , values ) {
    let text = I18n.get( key );
    Object.entries( values || {} ).forEach( ( [ name , value ] ) => {
      text = text.replace( new RegExp( "\\{" + name + "\\}" , "g" ) , () => String( value ) );
    } );
    return text;
  },

  label( account ) {
    return ( account.display_name || "#" + account.remote_user_id ) + " · " + ( account.project_name || "?" );
  },

  // One account as a line: the app, its name there, and what it is.
  account( account , extra ) {
    const badges = [];
    if ( account.owner ) badges.push( Dom.badge( I18n.get( "people.owner" ) , "dark" ) );
    if ( account.role ) badges.push( Dom.badge( I18n.get( "people.role_" + account.role ) || account.role , account.role === "admin" ? "primary" : "secondary" ) );
    if ( account.expired ) badges.push( Dom.badge( I18n.get( "people.expired" ) , "danger" ) );
    else if ( account.disabled ) badges.push( Dom.badge( I18n.get( "people.disabled" ) , "warning" ) );
    return Dom.el( "div" , { class: "d-flex flex-wrap align-items-center gap-1 py-1" , children: [
      Dom.el( "span" , { class: "text-body-secondary small" , text: account.project_name || "?" } ),
      Dom.el( "span" , { class: "fw-semibold text-break" , text: account.display_name || "#" + account.remote_user_id } ),
      ...badges,
      extra || null,
    ].filter( Boolean ) } );
  },

  accounts( accounts ) {
    return Dom.el( "div" , { children: accounts.map( ( account ) => this.account( account ) ) } );
  },

  // How many devices, and the most recent one.
  devices( accounts ) {
    const listed = accounts.filter( ( account ) => account.devices_listed );
    if ( !listed.length ) return Dom.el( "span" , { class: "small text-body-secondary" , text: I18n.get( "people.devices_unknown" ) } );
    const devices = listed.flatMap( ( account ) => account.devices || [] );
    if ( !devices.length ) return Dom.el( "span" , { class: "small text-body-secondary" , text: I18n.get( "people.devices_none" ) } );
    const seen = ( device ) => new Date( device.last_seen_at || device.created_at ).getTime();
    const latest = devices.slice().sort( ( a , b ) => seen( b ) - seen( a ) )[ 0 ];
    const addresses = [ ...new Set( devices.map( ( device ) => device.ip ).filter( Boolean ) ) ];
    return Dom.el( "div" , { class: "small" , children: [
      Dom.el( "div" , { text: devices.length === 1 ? I18n.get( "people.devices_one" ) : this.words( "people.devices_count" , { count: devices.length } ) } ),
      Dom.el( "div" , { class: "text-body-secondary text-break" , children: [
        Dom.el( "span" , { text: ( latest.label || latest.user_agent || "" ).slice( 0 , 60 ) + " " } ),
        Dom.el( "span" , { text: I18n.get( "people.last_seen" ).split( "{date}" )[ 0 ] } ),
        Dom.when( latest.last_seen_at || latest.created_at ),
      ] } ),
      addresses.length ? Dom.el( "div" , { class: "text-body-secondary font-monospace text-break" , text: addresses.slice( 0 , 3 ).join( ", " ) } ) : null,
    ].filter( Boolean ) } );
  },

  button( key , handler , tone ) {
    return Dom.el( "button" , {
      class: "btn btn-sm " + ( tone || "btn-outline-secondary" ),
      attrs: { type: "button" },
      text: I18n.get( key ),
      on: { click: handler },
    } );
  },

  matches( text ) {
    const q = Dom.get( "people-search" ).value.trim().toLowerCase();
    return !q || text.toLowerCase().includes( q );
  },

  fill( tableId , rows ) {
    const table = Dom.get( tableId );
    Dom.swap( table.querySelector( "tbody" ) , ...rows );
    if ( rows.length === 0 ) Pager.empty( table );
    Pager.label( tableId );
  },

  // --- the screen ------------------------------------------------------------

  render() {
    const data = this.data;
    if ( !data ) return;
    Dom.text( Dom.get( "people-synced" ) , data.synced_at
      ? this.words( "people.synced" , { date: new Date( data.synced_at ).toLocaleString() } ) : I18n.get( "people.never_synced" ) );

    this.fill( "people-suggestions-table" , data.suggestions.map( ( suggestion ) => Dom.el( "tr" , { children: [
      Dom.el( "td" , { children: [ this.accounts( suggestion.a ) ] } ),
      Dom.el( "td" , { children: [ this.accounts( suggestion.b ) ] } ),
      Dom.el( "td" , { children: [ Dom.el( "div" , { class: "d-flex flex-wrap gap-1" , children: suggestion.reasons.map(
        ( reason ) => Dom.badge( I18n.get( "people.reason_" + reason ) || reason , reason === "same_name" ? "info" : "secondary" ) ) } ) ] } ),
      Dom.el( "td" , { children: [ Dom.rowActions( this.label( suggestion.a[ 0 ] ) , [
        this.button( "people.confirm_button" , () => this.onConfirm( suggestion ) , "btn-outline-primary" ),
        this.button( "people.dismiss_button" , () => this.onDismiss( suggestion ) ),
      ] ) ] } ),
    ] } ) ) );

    const people = data.people.filter( ( person ) => this.matches( [ person.display_name , person.notes ,
      ...person.accounts.map( ( account ) => this.label( account ) ) ].join( " " ) ) );
    this.fill( "people-table" , people.map( ( person ) => Dom.el( "tr" , { children: [
      Dom.el( "td" , { children: [
        Dom.el( "div" , { class: "fw-semibold text-break" , text: person.display_name } ),
        person.notes ? Dom.el( "div" , { class: "small text-body-secondary text-break" , text: person.notes } ) : null,
      ].filter( Boolean ) } ),
      Dom.el( "td" , { children: person.accounts.map( ( account ) => this.account( account , Dom.el( "button" , {
        class: "btn btn-sm btn-link link-secondary p-0 ms-1",
        attrs: { type: "button" , title: I18n.get( "people.split_button" ) ,
          "aria-label": this.words( "people.split_label" , { account: this.label( account ) , name: person.display_name } ) },
        children: [ Dom.el( "i" , { class: "bi bi-x-circle" , attrs: { "aria-hidden": "true" } } ) ],
        on: { click: () => this.onSplit( person , account ) },
      } ) ) ) } ),
      Dom.el( "td" , { children: [ this.devices( person.accounts ) ] } ),
      Dom.el( "td" , { children: [ Dom.rowActions( person.display_name , [
        this.button( "people.rename_button" , () => this.onRename( person ) ),
        data.people.length > 1 ? this.button( "people.merge_button" , () => this.onMerge( person ) ) : null,
      ] ) ] } ),
    ] } ) ) );

    const loose = data.unlinked.filter( ( account ) => this.matches( this.label( account ) ) );
    this.fill( "people-unlinked-table" , loose.map( ( account ) => Dom.el( "tr" , { children: [
      Dom.el( "td" , { children: [ this.account( account ) ] } ),
      Dom.el( "td" , { children: [ this.devices( [ account ] ) ] } ),
      Dom.el( "td" , { children: [ Dom.rowActions( this.label( account ) , [
        this.button( "people.new_person_button" , () => this.onLink( [ account ] , 0 ) ),
        data.people.length ? this.button( "people.add_to_button" , () => this.onAddTo( account ) ) : null,
      ] ) ] } ),
    ] } ) ) );

    this.fill( "people-service-table" , data.service.map( ( account ) => {
      const what = account.hand_made_portal
        ? Dom.badge( this.words( "people.hand_made" , { app: account.project_name || "?" } ) , "warning" )
        : Dom.badge( I18n.get( account.portal ? "people.this_portal" : "people.service" ) , account.portal ? "primary" : "info" );
      return Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ this.account( account ) ] } ),
        Dom.el( "td" , { children: [ what ] } ),
      ] } );
    } ) );
  },

  // --- actions ---------------------------------------------------------------

  async act( work , message ) {
    try {
      const result = await work();
      if ( message ) Dom.toast( typeof message === "function" ? message( result ) : message , "ok" );
    } catch ( error ) { App.showError( error.message ); }
    await this.refresh();
  },

  onLink( accounts , personId ) {
    return this.act( () => Api.linkPeople( accounts.map( ( account ) => this.ref( account ) ) , personId ) ,
      ( reply ) => this.words( "people.linked" , { name: reply.person.display_name } ) );
  },

  onConfirm( suggestion ) {
    return this.onLink( suggestion.a.concat( suggestion.b ) , suggestion.a_person_id || suggestion.b_person_id || 0 );
  },

  onDismiss( suggestion ) {
    return this.act( () => Api.dismissPeople( suggestion.a.map( ( account ) => this.ref( account ) ) , suggestion.b.map( ( account ) => this.ref( account ) ) ) );
  },

  async onSplit( person , account ) {
    if ( !await Dom.confirm( this.words( "people.split_confirm" , { account: this.label( account ) , name: person.display_name } ) ) ) return;
    await this.act( () => Api.unlinkPerson( this.ref( account ) ) );
  },

  async onRename( person ) {
    const saved = await Dom.form( {
      title: this.words( "people.rename_heading" , { name: person.display_name } ),
      fields: [
        { name: "name" , label: I18n.get( "people.name_label" ) , value: person.display_name , required: true , maxLength: 80 },
        { name: "notes" , label: I18n.get( "people.notes_label" ) , value: person.notes || "" , maxLength: 2000 },
      ],
      save: ( values ) => Api.editPerson( person.id , values.name , values.notes ),
    } );
    if ( saved !== null ) await this.refresh();
  },

  async onMerge( person ) {
    const others = this.data.people.filter( ( other ) => other.id !== person.id );
    const saved = await Dom.form( {
      title: this.words( "people.merge_heading" , { name: person.display_name } ),
      fields: [ { name: "into" , type: "select" , label: I18n.get( "people.merge_label" ) ,
        options: others.map( ( other ) => ( { value: String( other.id ) , label: other.display_name } ) ) , value: String( others[ 0 ].id ) } ],
      save: ( values ) => Api.mergePeople( Number( values.into ) , person.id ),
    } );
    if ( saved === null ) return;
    Dom.toast( I18n.get( "people.merged" ) , "ok" );
    await this.refresh();
  },

  async onAddTo( account ) {
    const people = this.data.people;
    const saved = await Dom.form( {
      title: this.words( "people.add_to_heading" , { name: this.label( account ) } ),
      fields: [ { name: "person" , type: "select" , label: I18n.get( "people.add_to_label" ) ,
        options: people.map( ( person ) => ( { value: String( person.id ) , label: person.display_name } ) ) , value: String( people[ 0 ].id ) } ],
      save: ( values ) => Api.linkPeople( [ this.ref( account ) ] , Number( values.person ) ),
    } );
    if ( saved !== null ) await this.refresh();
  },
};
