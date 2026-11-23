// The Access screen (admins): one project at a time, everyone in the portal
// against it. Every project is off for everyone by default; switching
// someone on lets them see and join it. showPerson is the same list turned
// around: every project against one person, under the Users table. From here an admin can also put someone on
// a project now, match them to an account the project already had, or take
// them off it here. app.js calls Access.init( me ) for admins.
const Access = {
  projectId: null,
  projects: [],

  async init( me ) {
    if ( me.role !== "admin" ) return;
    Dom.show( Dom.get( "access-panel" ) , true );
    Dom.get( "access-project" ).addEventListener( "change" , ( event ) => {
      this.projectId = Number( event.target.value ) || null;
      Pager.reset( "access-table" );
      this.refresh();
    } );
    Pager.sortable( "access-table" , () => this.refresh() );
    Pager.search( "access-table" , "access-search" , () => this.refresh() );
    Pager.sortable( "user-projects-table" , () => this.refreshPerson() );
    Pager.search( "user-projects-table" , "user-projects-search" , () => this.refreshPerson() );
    Dom.get( "user-projects-close" ).addEventListener( "click" , () => this.closePerson() );
    await this.fillProjects();
  },

  // The project picker, keeping the one picked while it still exists.
  async fillProjects() {
    const page = await Api.listCatalog( 0 , { limit: 1000 } );
    this.projects = page.projects;
    const select = Dom.get( "access-project" );
    const options = page.projects.map( ( project ) => Dom.el( "option" , { text: project.name , attrs: { value: String( project.id ) } } ) );
    Dom.swap( select , ...options );
    const ids = page.projects.map( ( project ) => project.id );
    if ( !ids.includes( this.projectId ) ) this.projectId = ids.length ? ids[ 0 ] : null;
    if ( this.projectId ) select.value = String( this.projectId );
    Dom.show( Dom.get( "access-list-card" ) , this.projectId !== null );
    Dom.show( Dom.get( "access-none" ) , this.projectId === null );
    if ( this.projectId ) await this.refresh();
  },

  // From a Catalog row's Access button.
  async showFor( projectId ) {
    this.projectId = projectId;
    Nav.show( "access" );
    await this.fillProjects();
  },

  status( row ) {
    if ( !row.member ) return Dom.el( "span" , { class: "text-body-secondary" , text: I18n.get( "access.not_member" ) } );
    const via = I18n.get( "access.via_" + row.via );
    const name = row.remote_name ? I18n.get( "access.as_name" ).replace( "{name}" , () => row.remote_name ) : "";
    return Dom.el( "div" , { children: [
      Dom.badge( via , row.via === "own_key" ? "info" : "success" ),
      name ? Dom.el( "div" , { class: "small text-body-secondary" , text: name } ) : null,
    ].filter( Boolean ) } );
  },

  allowSwitch( row ) {
    const id = "access-allow-" + row.user_id;
    const input = Dom.el( "input" , { class: "form-check-input" , attrs: { type: "checkbox" , role: "switch" , id: id } } );
    input.checked = row.allowed;
    if ( row.role === "admin" ) {
      input.disabled = true;
      input.title = I18n.get( "access.admin_always" );
    }
    input.addEventListener( "change" , () => this.onAllow( row , input ) );
    return Dom.el( "div" , { class: "form-check form-switch mb-0" , children: [
      input,
      Dom.el( "label" , { class: "visually-hidden" , text: I18n.get( "access.allow_label" ).replace( "{name}" , () => row.display_name ) , attrs: { for: id } } ),
    ] } );
  },

  async refresh() {
    if ( !this.projectId ) return;
    const projectId = this.projectId;
    const body = Dom.get( "access-table" ).querySelector( "tbody" );
    await Pager.load( "access-table" , "rows" , ( offset ) => Api.projectAccess( projectId , offset , Pager.params( "access-table" ) ) , ( row ) => {
      const nameCell = Dom.el( "td" , { children: [
        Dom.el( "div" , { class: "fw-semibold" , text: row.display_name } ),
        row.role === "admin" ? Dom.el( "div" , { class: "small text-body-secondary" , text: I18n.get( "admin.role_admin" ) } ) : null,
      ].filter( Boolean ) } );
      body.appendChild( Dom.el( "tr" , { children: [
        nameCell,
        Dom.el( "td" , { children: [ this.allowSwitch( row ) ] } ),
        Dom.el( "td" , { children: [ this.status( row ) ] } ),
        Dom.el( "td" , { children: [ this.actions( row ) ] } ),
      ] } ) );
    } );
  },

  actions( row ) {
    const button = ( key , handler , danger ) => Dom.el( "button" , {
      class: "btn btn-sm " + ( danger ? "btn-outline-danger" : "btn-outline-secondary" ),
      attrs: { type: "button" },
      text: I18n.get( key ),
      on: { click: handler },
    } );
    return Dom.rowActions( row.display_name , [
      row.member || !row.allowed ? null : button( "access.grant_button" , () => this.onGrant( row ) ),
      button( "access.link_button" , () => this.onLink( row ) ),
      row.member ? button( "access.remove_button" , () => this.onRemove( row ) , true ) : null,
    ] );
  },

  async onAllow( row , input ) {
    const projectId = this.projectId;
    const allowed = input.checked;
    let onProject = "keep";
    if ( !allowed && row.member ) {
      onProject = await this.onProjectChoice( row , I18n.get( "access.deny_heading" ).replace( "{name}" , () => row.display_name ) );
      if ( onProject === null ) {
        input.checked = true;
        return;
      }
    }
    // Undo lets them back in here; it does not switch an account back on there.
    const set = async ( value ) => {
      await Api.setProjectAccess( projectId , row.user_id , value , value ? "keep" : onProject );
      if ( this.projectId === projectId ) await this.refresh();
    };
    try {
      await Undo.run( {
        label: I18n.get( allowed ? "access.undo_allow" : "access.undo_deny" ).replace( "{name}" , () => row.display_name ),
        do: () => set( allowed ),
        undo: () => set( !allowed ),
      } );
    } catch ( error ) {
      input.checked = !allowed;
      App.showError( error.message );
    }
  },

  // One person's apps, from their Apps button on the Users table.
  async showPerson( user ) {
    if ( !this.person || this.person.id !== user.id ) {
      Pager.reset( "user-projects-table" );
      Dom.get( "user-projects-search" ).value = "";
    }
    this.person = user;
    Dom.text( Dom.get( "user-projects-heading" ) , I18n.get( "access.person_heading" ).replace( "{name}" , () => user.display_name ) );
    Dom.show( Dom.get( "user-projects" ) , true );
    await this.refreshPerson();
    Dom.reveal( Dom.get( "user-projects" ) );
  },

  closePerson() {
    this.person = null;
    Dom.show( Dom.get( "user-projects" ) , false );
  },

  async refreshPerson() {
    const user = this.person;
    if ( !user ) return;
    const body = Dom.get( "user-projects-table" ).querySelector( "tbody" );
    await Pager.load( "user-projects-table" , "rows" , ( offset ) => Api.userProjects( user.id , offset , Pager.params( "user-projects-table" ) ) , ( row ) => {
      const nameCell = Dom.el( "td" , { children: [
        Dom.el( "div" , { class: "d-flex align-items-center gap-2" , children: [
          Projects.icon( row ),
          Dom.el( "div" , { children: [
            Dom.el( "div" , { class: "fw-semibold" , text: row.name } ),
            row.listed ? null : Dom.el( "div" , { class: "small text-body-secondary" , text: I18n.get( "access.hidden" ) } ),
          ].filter( Boolean ) } ),
        ] } ),
      ] } );
      const id = "user-project-allow-" + row.project_id;
      const input = Dom.el( "input" , { class: "form-check-input" , attrs: { type: "checkbox" , role: "switch" , id: id } } );
      input.checked = row.allowed;
      input.addEventListener( "change" , () => this.onPersonAllow( user , row , input ) );
      const allowCell = Dom.el( "td" , { children: [ Dom.el( "div" , { class: "form-check form-switch mb-0" , children: [
        input,
        Dom.el( "label" , { class: "visually-hidden" , text: I18n.get( "access.person_allow_label" ).replace( "{name}" , () => user.display_name ).replace( "{project}" , () => row.name ) , attrs: { for: id } } ),
      ] } ) ] } );
      body.appendChild( Dom.el( "tr" , { children: [
        nameCell,
        allowCell,
        Dom.el( "td" , { children: [ this.status( row ) ] } ),
      ] } ) );
    } );
  },

  async onPersonAllow( user , row , input ) {
    const allowed = input.checked;
    let onProject = "keep";
    if ( !allowed && row.member ) {
      // onProjectChoice reads the row the Access screen's way.
      const asAccess = { ...row , display_name: user.display_name };
      onProject = await this.onProjectChoice( asAccess , I18n.get( "access.person_deny_heading" ).replace( "{name}" , () => user.display_name ).replace( "{project}" , () => row.name ) );
      if ( onProject === null ) {
        input.checked = true;
        return;
      }
    }
    const set = async ( value ) => {
      await Api.setProjectAccess( row.project_id , user.id , value , value ? "keep" : onProject );
      if ( this.person && this.person.id === user.id ) await this.refreshPerson();
      if ( this.projectId === row.project_id ) await this.refresh();
    };
    try {
      await Undo.run( {
        label: I18n.get( allowed ? "access.undo_allow" : "access.undo_deny" ).replace( "{name}" , () => user.display_name ),
        do: () => set( allowed ),
        undo: () => set( !allowed ),
      } );
    } catch ( error ) {
      input.checked = !allowed;
      App.showError( error.message );
    }
  },

  project() {
    return this.projects.find( ( project ) => project.id === this.projectId ) || { capabilities: [] };
  },

  has( capability ) {
    return ( this.project().capabilities || [] ).includes( capability );
  },

  // Putting someone on: straight away on an app without view-only or
  // expiring accounts, otherwise asking which.
  async onGrant( row ) {
    const fields = [];
    if ( this.has( "viewer_role" ) ) {
      fields.push( { name: "role" , type: "select" , label: I18n.get( "manage.role_label" ) , value: "user" ,
        options: [ "user" , "viewer" ].map( ( role ) => ( { value: role , label: I18n.get( "manage.role_" + role ) } ) ) } );
    }
    if ( this.has( "account_expiry" ) ) {
      fields.push( { name: "expires_in_days" , type: "number" , min: 0 , max: 3650 , label: I18n.get( "manage.days_label" ) ,
        help: I18n.get( "manage.expiry_help" ) , placeholder: I18n.get( "manage.days_never" ) } );
    }
    const grant = ( values ) => Api.grantProject( this.projectId , row.user_id , {
      role: values.role || "user" , expires_in_days: values.expires_in_days || 0 ,
    } );
    try {
      if ( fields.length === 0 ) {
        await grant( {} );
      } else {
        const saved = await Dom.form( {
          title: I18n.get( "access.grant_heading" ).replace( "{name}" , () => row.display_name ),
          submit: I18n.get( "access.grant_button" ),
          fields: fields,
          save: grant,
        } );
        if ( saved === null ) return;
      }
      Dom.toast( I18n.get( "access.granted" ).replace( "{name}" , () => row.display_name ) , "ok" );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
  },

  // What happens to someone's account on the project as they come off it:
  // asked when the portal has an account there it could act on. Answers
  // "keep", "disable", "revoke", or null for Cancel.
  async onProjectChoice( row , title ) {
    if ( !row.member || !row.remote_user_id || row.via === "own_key" ) {
      return ( await Dom.confirm( title ) ) ? "keep" : null;
    }
    const saved = await Dom.form( {
      title: title,
      submit: I18n.get( "access.remove_go" ),
      fields: [ { name: "on_project" , type: "select" , label: I18n.get( "access.on_project_label" ) ,
        help: I18n.get( "access.on_project_help" ) , value: "disable" ,
        options: [ "disable" , "revoke" , "keep" ].map( ( choice ) => ( { value: choice , label: I18n.get( "access.on_project_" + choice ) } ) ) } ],
      save: ( values ) => values.on_project,
    } );
    return saved;
  },

  // Which of the project's own accounts is theirs: picked from the project's
  // list, which the portal fetches with its admin key.
  async onLink( row ) {
    let users;
    try {
      users = ( await Api.remoteUsers( this.projectId ) ).users;
    } catch ( error ) {
      App.showError( error.message );
      return;
    }
    // Never the owner, the portal's own account or another service account: nobody signs into those.
    const options = users.filter( ( user ) => !user.owner && !user.portal && user.kind !== "service" ).map( ( user ) => ( {
      value: String( user.id ),
      label: user.display_name + " (" + user.role + ( user.disabled ? ", " + I18n.get( "access.disabled" ) : "" ) + ")",
    } ) );
    if ( options.length === 0 ) {
      App.showError( I18n.get( "access.no_remote_users" ) );
      return;
    }
    const picked = options.find( ( option ) => option.label.toLowerCase().startsWith( row.display_name.toLowerCase() + " (" ) );
    const saved = await Dom.form( {
      title: I18n.get( "access.link_heading" ).replace( "{name}" , () => row.display_name ),
      fields: [ { name: "remote_user_id" , type: "select" , label: I18n.get( "access.link_label" ) ,
        help: I18n.get( "access.link_help" ) , options: options , value: picked ? picked.value : options[ 0 ].value } ],
      save: ( values ) => Api.linkRemoteUser( this.projectId , row.user_id , Number( values.remote_user_id ) ),
    } );
    if ( saved === null ) return;
    await this.refresh();
  },

  async onRemove( row ) {
    const onProject = await this.onProjectChoice( row , I18n.get( "access.remove_heading" ).replace( "{name}" , () => row.display_name ) );
    if ( onProject === null ) return;
    try {
      await Api.removeProjectMember( this.projectId , row.user_id , onProject );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
  },
};
