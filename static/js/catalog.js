// A project's icon, picked from a short grid instead of typed: nobody knows
// Bootstrap Icons names. An icon set before the grid existed that is not in it
// is kept as one more choice, so editing a project never changes it unasked.
const IconPicker = {
  fallback: "kanban",
  choices: [
    "kanban", "list-check", "clipboard-data", "journal-text", "newspaper", "book",
    "chat-dots", "envelope", "calendar-event", "people", "briefcase", "house-door",
    "cart3", "wallet2", "graph-up-arrow", "box-seam", "map", "airplane",
    "camera", "film", "music-note-beamed", "palette", "controller", "cup-hot",
    "heart-pulse", "lightbulb", "tools", "code-slash", "terminal", "database",
    "cloud", "shield-lock", "robot", "rocket-takeoff", "globe2", "app-indicator",
  ],

  build( id , value ) {
    const names = this.choices.includes( value ) || !/^[a-z0-9-]+$/.test( value || "" ) ? this.choices : this.choices.concat( [ value ] );
    const picked = names.includes( value ) ? value : this.fallback;
    const radios = [];
    const options = names.map( ( name , index ) => {
      const radioId = id + "-" + index;
      const radio = Dom.el( "input" , { class: "btn-check" , attrs: { type: "radio" , name: id , id: radioId , value: name , autocomplete: "off" } } );
      // defaultChecked too, so the Add form's reset goes back to it.
      radio.checked = radio.defaultChecked = name === picked;
      radios.push( radio );
      const words = name.replace( /-/g , " " );
      return Dom.el( "div" , { children: [
        radio,
        Dom.el( "label" , { class: "btn btn-outline-secondary icon-choice" , attrs: { for: radioId , title: words , "aria-label": words } ,
          children: [ Dom.el( "i" , { class: "bi bi-" + name , attrs: { "aria-hidden": "true" } } ) ] } ),
      ] } );
    } );
    const element = Dom.el( "fieldset" , { children: [
      Dom.el( "legend" , { class: "form-label fs-6" , text: I18n.get( "catalog.icon_label" ) } ),
      Dom.el( "div" , { class: "d-flex flex-wrap gap-1" , children: options } ),
    ] } );
    return { element , read: () => ( radios.find( ( radio ) => radio.checked ) || {} ).value || this.fallback };
  },
};

// The Catalog screen (admins): the projects the portal knows, each with the
// admin key it uses there. Adding one with its key is all it takes for
// everyone to see it and join it. app.js calls Catalog.init( me ) for admins;
// the server checks again on every call.
const Catalog = {
  me: null,

  async init( me ) {
    if ( me.role !== "admin" ) return;
    this.me = me;
    Dom.show( Dom.get( "catalog-panel" ) , true );
    this.addIcon = IconPicker.build( "project-icon" , IconPicker.fallback );
    Dom.get( "project-icon-picker" ).replaceChildren( this.addIcon.element );
    Dom.onSubmit( Dom.get( "add-project-form" ) , this.onAdd.bind( this ) );
    Pager.sortable( "catalog-table" , () => this.refresh() );
    Pager.search( "catalog-table" , "catalog-search" , () => this.refresh() );
    window.addEventListener( "nav" , () => {
      if ( Nav.current() === "catalog" ) this.refresh();
    } );
    await this.refresh();
  },

  async onAdd( event ) {
    event.preventDefault();
    const form = event.target;
    const value = ( id ) => Dom.get( id ).value.trim();
    try {
      const created = await Api.createProject( {
        name: value( "project-name" ),
        base_url: value( "project-url" ),
        admin_key: value( "project-key" ),
        description: value( "project-description" ),
        icon: this.addIcon.read(),
        color: Dom.get( "project-color-set" ).checked ? Dom.get( "project-color" ).value : "",
        listed: Dom.get( "project-listed" ).checked,
        github_url: value( "project-github" ),
        board_url: value( "project-board" ),
        board_share_url: value( "project-share" ),
      } );
      form.reset();
      Dom.get( "project-listed" ).checked = true;
      Undo.saved( form );
      Dom.toast( I18n.get( "catalog.added" ).replace( "{name}" , () => created.project.name ) , "ok" );
      await this.refresh();
      if ( typeof Access !== "undefined" ) await Access.fillProjects();
    } catch ( error ) {
      App.showError( error.message );
    }
  },

  keyBadge( project ) {
    const tones = { ok: "success" , invalid: "danger" , not_admin: "warning" , missing: "secondary" };
    const status = project.key_set ? project.key_status || "ok" : "missing";
    return Dom.badge( I18n.get( "catalog.key_" + status ) , tones[ status ] || "secondary" );
  },

  healthBadge( project ) {
    if ( !project.health ) return Dom.badge( I18n.get( "catalog.health_unknown" ) , "secondary" );
    const badge = Dom.badge( I18n.get( "catalog.health_" + project.health ) , project.health === "up" ? "success" : "danger" );
    if ( project.check_note ) badge.title = project.check_note;
    return badge;
  },

  async refresh() {
    const body = Dom.get( "catalog-table" ).querySelector( "tbody" );
    await Pager.load( "catalog-table" , "projects" , ( offset ) => Api.listCatalog( offset , Pager.params( "catalog-table" ) ) , ( project ) => {
      const name = Dom.el( "div" , { class: "d-flex align-items-center gap-2" , children: [
        Projects.icon( project ),
        Dom.el( "div" , { children: [
          Dom.el( "div" , { class: "fw-semibold" , text: project.name } ),
          project.listed ? null : Dom.el( "div" , { class: "small text-body-secondary" , text: I18n.get( "catalog.hidden" ) } ),
        ].filter( Boolean ) } ),
      ] } );
      const link = Dom.el( "div" , { children: [
        Dom.el( "a" , { class: "text-break" , text: project.base_url , attrs: { href: project.base_url , target: "_blank" , rel: "noopener noreferrer" } } ),
        this.links( project ),
      ].filter( Boolean ) } );
      // Why the last check was unhappy, under the status it explains.
      const note = project.check_note ? Dom.el( "div" , { class: "small text-body-secondary text-break" , text: project.check_note } ) : null;
      const down = project.health === "down";
      body.appendChild( Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ name ] } ),
        Dom.el( "td" , { children: [ link ] } ),
        Dom.el( "td" , { children: [ this.keyBadge( project ) , down ? null : note , this.contract( project ) ].filter( Boolean ) } ),
        Dom.el( "td" , { children: [ this.healthBadge( project ) , down ? note : null , Dom.el( "div" , { class: "small text-body-secondary" , children: [ Dom.when( project.checked_at , I18n.get( "catalog.never_checked" ) ) ] } ) ].filter( Boolean ) } ),
        Dom.el( "td" , { text: String( project.members ) } ),
        Dom.el( "td" , { children: [ this.actions( project ) ] } ),
      ] } ) );
    } );
  },

  // Whether Manage project can do everything there (portal contract v2), or
  // which upgrade it is waiting on. Nothing before the first good check.
  contract( project ) {
    if ( project.key_status !== "ok" ) return null;
    const all = [ "viewer_role" , "account_expiry" , "invite_links" , "key_minting" , "key_sign_in" ];
    const missing = all.filter( ( capability ) => !( project.capabilities || [] ).includes( capability ) );
    const text = missing.length === 0 ? I18n.get( "catalog.contract_ok" )
      : I18n.get( "catalog.contract_behind" ).replace( "{missing}" , () => missing.map( ( capability ) => I18n.get( "manage.cap_" + capability ) ).join( ", " ) );
    return Dom.el( "div" , { class: "small " + ( missing.length ? "text-warning-emphasis" : "text-body-secondary" ) , text: text } );
  },

  // GitHub, the board and its view-only link, as small icons under the address.
  links( project ) {
    const items = [
      [ project.github_url , "bi-github" , "catalog.link_github" ],
      [ project.board_url , "bi-kanban" , "catalog.link_board" ],
      [ project.board_share_url , "bi-eye" , "catalog.link_share" ],
    ].filter( ( [ href ] ) => href );
    if ( !items.length ) return null;
    return Dom.el( "div" , { class: "d-flex gap-3 mt-1" , children: items.map( ( [ href , icon , key ] ) => Dom.el( "a" , {
      class: "link-secondary",
      attrs: { href , target: "_blank" , rel: "noopener noreferrer" , title: I18n.get( key ) , "aria-label": I18n.get( key ) + ": " + project.name },
      children: [ Dom.el( "i" , { class: "bi " + icon , attrs: { "aria-hidden": "true" } } ) ],
    } ) ) } );
  },

  actions( project ) {
    const button = ( key , handler , danger ) => Dom.el( "button" , {
      class: "btn btn-sm " + ( danger ? "btn-outline-danger" : "btn-outline-secondary" ),
      attrs: { type: "button" },
      text: I18n.get( key ),
      on: { click: handler },
    } );
    return Dom.rowActions( project.name , [
      button( "catalog.check_button" , () => this.onCheck( project ) ),
      button( "catalog.manage_button" , () => Manage.showFor( project.id ) ),
      button( "catalog.access_button" , () => Access.showFor( project.id ) ),
      button( "catalog.edit_button" , () => this.onEdit( project ) ),
      button( "catalog.key_button" , () => this.onKey( project ) ),
      button( "catalog.delete_button" , () => this.onDelete( project ) , true ),
    ] );
  },

  async onCheck( project ) {
    try {
      const checked = await Api.checkProject( project.id );
      const status = checked.project.health === "up" && checked.project.key_status === "ok" ? "catalog.check_ok" : "catalog.check_problem";
      Dom.toast( I18n.get( status ).replace( "{name}" , () => project.name ) , status === "catalog.check_ok" ? "ok" : "error" );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
  },

  fields( project ) {
    return {
      name: project.name , base_url: project.base_url , description: project.description ,
      icon: project.icon , color: project.color , listed: project.listed ,
      github_url: project.github_url || "" , board_url: project.board_url || "" , board_share_url: project.board_share_url || "",
    };
  },

  async onEdit( project ) {
    const before = this.fields( project );
    const saved = await Dom.form( {
      title: I18n.get( "catalog.edit_heading" ).replace( "{name}" , () => project.name ),
      fields: [
        { name: "name" , label: I18n.get( "catalog.name_label" ) , value: before.name , required: true , maxLength: 80 },
        { name: "base_url" , type: "url" , label: I18n.get( "catalog.url_label" ) , value: before.base_url , required: true , maxLength: 300 , help: I18n.get( "catalog.url_change_help" ) },
        { name: "description" , label: I18n.get( "catalog.description_label" ) , value: before.description , maxLength: 500 },
        { name: "icon" , type: "custom" , build: ( id ) => IconPicker.build( id , before.icon || IconPicker.fallback ) },
        { name: "color" , label: I18n.get( "catalog.color_label" ) , value: before.color , maxLength: 7 , placeholder: "#0d6efd" },
        { name: "github_url" , type: "url" , label: I18n.get( "catalog.github_label" ) , value: before.github_url , maxLength: 300 , placeholder: I18n.get( "catalog.github_placeholder" ) },
        { name: "board_url" , type: "url" , label: I18n.get( "catalog.board_label" ) , value: before.board_url , maxLength: 300 , placeholder: I18n.get( "catalog.board_placeholder" ) , help: I18n.get( "catalog.links_help" ) },
        { name: "board_share_url" , type: "url" , label: I18n.get( "catalog.share_label" ) , value: before.board_share_url , maxLength: 500 , help: I18n.get( "catalog.share_help" ) },
        { name: "listed" , type: "checkbox" , label: I18n.get( "catalog.listed_label" ) , value: before.listed },
      ],
      save: ( values ) => Api.editProject( project.id , values ),
    } );
    if ( saved === null ) return;
    if ( saved.keys_dropped ) Dom.toast( I18n.get( "catalog.keys_dropped" ) , "error" );
    // A changed address dropped the keys, which undo cannot bring back, so
    // only an edit that kept them is undoable.
    if ( !saved.keys_dropped ) {
      const after = this.fields( saved.project );
      Undo.record( {
        label: I18n.get( "catalog.undo_edit" ),
        undo: async () => { await Api.editProject( project.id , before ); await this.refresh(); },
        redo: async () => { await Api.editProject( project.id , after ); await this.refresh(); },
      } );
    }
    await this.refresh();
  },

  async onKey( project ) {
    const saved = await Dom.form( {
      title: I18n.get( "catalog.key_heading" ).replace( "{name}" , () => project.name ),
      fields: [ { name: "admin_key" , type: "password" , required: true , maxLength: 400 ,
        label: I18n.get( "catalog.key_label" ) , help: I18n.get( "catalog.key_help" ) } ],
      save: ( values ) => Api.setProjectKey( project.id , values.admin_key ),
    } );
    if ( saved === null ) return;
    Dom.toast( I18n.get( "catalog.key_saved" ) , "ok" );
    await this.refresh();
  },

  async onDelete( project ) {
    try {
      await Undo.run( {
        label: I18n.get( "catalog.undo_delete" ).replace( "{name}" , () => project.name ),
        toast: true,
        do: async () => { await Api.deleteProject( project.id ); await this.refresh(); },
        undo: async () => { await Api.restoreProject( project.id ); await this.refresh(); },
      } );
      if ( typeof Access !== "undefined" ) await Access.fillProjects();
    } catch ( error ) { App.showError( error.message ); }
  },
};
