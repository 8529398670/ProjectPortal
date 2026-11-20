// The Access screen (admins): one project at a time, everyone in the portal
// against it. Every project is open to everyone by default; switching
// someone off keeps it from them. From here an admin can also put someone on
// a project now, match them to an account the project already had, or take
// them off it here. app.js calls Access.init( me ) for admins.
const Access = {
  projectId: null,

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
    await this.fillProjects();
  },

  // The project picker, keeping the one picked while it still exists.
  async fillProjects() {
    const page = await Api.listCatalog( 0 , { limit: 1000 } );
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
    const set = async ( allowed ) => {
      await Api.setProjectAccess( projectId , row.user_id , allowed );
      if ( this.projectId === projectId ) await this.refresh();
    };
    const allowed = input.checked;
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

  async onGrant( row ) {
    try {
      await Api.grantProject( this.projectId , row.user_id );
      Dom.toast( I18n.get( "access.granted" ).replace( "{name}" , () => row.display_name ) , "ok" );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
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
    const options = users.filter( ( user ) => !user.owner ).map( ( user ) => ( {
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
    if ( !( await Dom.confirm( I18n.get( "access.remove_confirm" ).replace( "{name}" , () => row.display_name ) ) ) ) return;
    try {
      await Api.removeProjectMember( this.projectId , row.user_id );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
  },
};
