// The Projects screen: every project you may use, as tiles. The ones you are
// on come first (recently opened first, or by name) with an Open button, then the
// ones you could join. Opening asks the portal for a fresh login link to that
// project and goes there, which is how anyone signed out of a project gets
// back in. app.js calls Projects.init( me ) once it knows who is signed in.
//
// The list is small (a portal has tens of projects, not thousands), so it is
// loaded whole and searched here, which keeps typing in the search box instant.
const Projects = {
  me: null,
  all: [],

  async init( me ) {
    this.me = me;
    Dom.typed( Dom.get( "projects-search" ) , () => this.render() );
    Dom.get( "projects-tag" ).addEventListener( "change" , () => this.render() );
    Dom.get( "projects-favorites" ).addEventListener( "change" , () => this.render() );
    // The sort is remembered in this browser only; nothing breaks without it.
    const sort = Dom.get( "projects-sort" );
    try { if ( localStorage.getItem( "projects-sort" ) === "name" ) sort.value = "name"; } catch ( error ) { /* storage off */ }
    sort.addEventListener( "change" , () => {
      try { localStorage.setItem( "projects-sort" , sort.value ); } catch ( error ) { /* storage off */ }
      this.render();
    } );
    // Back on the screen (from Catalog, say): the list may have changed.
    window.addEventListener( "nav" , () => {
      if ( Nav.current() === "projects" ) this.refresh();
    } );
    await this.refresh();
  },

  async refresh() {
    const rows = [];
    let offset = 0;
    for ( ;; ) {
      const page = await Api.listProjects( offset );
      rows.push( ...page.projects );
      if ( page.next === null || page.next === undefined ) break;
      offset = page.next;
    }
    this.all = rows;
    this.fillTags();
    this.render();
  },

  // The tag filter offers every tag in use, and keeps the one picked.
  fillTags() {
    const select = Dom.get( "projects-tag" );
    const picked = select.value;
    const tags = [ ...new Set( this.all.flatMap( ( project ) => project.tags ) ) ].sort();
    const options = [ Dom.el( "option" , { text: I18n.get( "projects.all_tags" ) , attrs: { value: "" } } ) ];
    tags.forEach( ( tag ) => options.push( Dom.el( "option" , { text: tag , attrs: { value: tag } } ) ) );
    Dom.swap( select , ...options );
    select.value = tags.includes( picked ) ? picked : "";
    Dom.show( Dom.get( "projects-tag-field" ) , tags.length > 0 );
  },

  matches( project ) {
    const query = Dom.get( "projects-search" ).value.trim().toLowerCase();
    const tag = Dom.get( "projects-tag" ).value;
    const favorites = Dom.get( "projects-favorites" ).checked;
    if ( favorites && !project.favorite ) return false;
    if ( tag !== "" && !project.tags.includes( tag ) ) return false;
    if ( query === "" ) return true;
    return ( project.name + " " + project.description + " " + project.tags.join( " " ) ).toLowerCase().includes( query );
  },

  // Recently opened first (never opened last, by name), or by name alone.
  sorted( projects ) {
    const byName = ( a , b ) => a.name.localeCompare( b.name , undefined , { sensitivity: "base" } );
    if ( Dom.get( "projects-sort" ).value === "name" ) return projects.slice().sort( byName );
    const opened = ( project ) => project.last_opened_at ? Date.parse( project.last_opened_at ) : 0;
    return projects.slice().sort( ( a , b ) => ( opened( b ) - opened( a ) ) || byName( a , b ) );
  },

  render() {
    const shown = this.sorted( this.all.filter( ( project ) => this.matches( project ) ) );
    const mine = shown.filter( ( project ) => project.member );
    const available = shown.filter( ( project ) => !project.member );
    Dom.swap( Dom.get( "my-projects" ) , ...mine.map( ( project ) => this.tile( project ) ) );
    Dom.swap( Dom.get( "available-projects" ) , ...available.map( ( project ) => this.tile( project ) ) );
    const narrowed = shown.length !== this.all.length;
    const mineEmpty = Dom.get( "my-projects-empty" );
    const emptyKey = narrowed ? "ui.no_matches" : this.all.length === 0 ? "projects.none" : "projects.mine_empty";
    Dom.text( mineEmpty , I18n.get( emptyKey ) );
    Dom.show( mineEmpty , mine.length === 0 );
    Dom.show( Dom.get( "available-projects-card" ) , available.length > 0 );
  },

  icon( project ) {
    const name = /^[a-z0-9-]+$/.test( project.icon || "" ) ? project.icon : "app-indicator";
    const holder = Dom.el( "span" , {
      class: "project-icon",
      attrs: { "aria-hidden": "true" },
      children: [ Dom.el( "i" , { class: "bi bi-" + name } ) ],
    } );
    if ( /^#[0-9a-fA-F]{6}$/.test( project.color || "" ) ) holder.style.setProperty( "--project-color" , project.color );
    return holder;
  },

  health( project ) {
    if ( !project.health ) return null;
    const up = project.health === "up";
    const text = I18n.get( up ? "projects.health_up" : "projects.health_down" );
    return Dom.el( "span" , {
      class: "project-health " + ( up ? "text-success" : "text-danger" ),
      attrs: { title: text },
      children: [
        Dom.el( "i" , { class: "bi " + ( up ? "bi-circle-fill" : "bi-exclamation-circle-fill" ) , attrs: { "aria-hidden": "true" } } ),
        Dom.el( "span" , { class: "visually-hidden" , text: text } ),
      ],
    } );
  },

  tile( project ) {
    const badges = [];
    if ( project.new && !project.member ) badges.push( Dom.badge( I18n.get( "projects.badge_new" ) , "primary" ) );
    if ( !project.listed ) badges.push( Dom.badge( I18n.get( "projects.badge_hidden" ) , "secondary" ) );
    if ( project.own_key ) badges.push( Dom.badge( I18n.get( "projects.badge_own_key" ) , "info" ) );

    const title = Dom.el( "div" , { class: "d-flex align-items-start gap-2" , children: [
      this.icon( project ),
      Dom.el( "div" , { class: "flex-grow-1 min-w-0" , children: [
        Dom.el( "h4" , { class: "h6 mb-1 text-break" , text: project.name } ),
        Dom.el( "div" , { class: "small text-body-secondary text-break" , text: project.base_url.replace( /^https?:\/\// , "" ) } ),
      ] } ),
      this.health( project ),
    ] } );

    const body = [ title ];
    if ( badges.length ) body.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-1 mt-2" , children: badges } ) );
    if ( project.description ) body.push( Dom.el( "p" , { class: "small mt-2 mb-0" , text: project.description } ) );
    if ( project.member && project.remote_name ) {
      body.push( Dom.el( "p" , { class: "small text-body-secondary mt-2 mb-0" ,
        text: I18n.get( "projects.signed_in_as" ).replace( "{name}" , () => project.remote_name ) } ) );
    }
    if ( project.tags.length ) {
      body.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-1 mt-2" , children:
        project.tags.map( ( tag ) => Dom.el( "span" , { class: "badge rounded-pill text-bg-light border" , text: tag } ) ) } ) );
    }

    const actions = project.member ? this.memberActions( project ) : this.joinActions( project );
    return Dom.el( "div" , { class: "col" , children: [
      Dom.el( "article" , { class: "card h-100 project-tile" , attrs: { "aria-label": project.name } , children: [
        Dom.el( "div" , { class: "card-body" , children: body } ),
        Dom.el( "div" , { class: "card-footer bg-transparent d-flex flex-wrap align-items-center gap-1 gap-sm-2" , children: actions } ),
      ] } ),
    ] } );
  },

  joinActions( project ) {
    const join = Dom.el( "button" , {
      class: "btn btn-primary btn-sm",
      attrs: { type: "button" },
      children: [ Dom.el( "i" , { class: "bi bi-plus-lg me-1" , attrs: { "aria-hidden": "true" } } ) , Dom.el( "span" , { text: I18n.get( "projects.join_button" ) } ) ],
      on: { click: ( event ) => this.go( event , () => Api.joinProject( project.id ) , event.ctrlKey || event.metaKey || event.shiftKey ) },
    } );
    if ( !project.ready ) {
      join.classList.add( "disabled" );
      join.setAttribute( "aria-disabled" , "true" );
      join.title = I18n.get( "projects.not_ready" );
    }
    const extra = Dom.el( "button" , {
      class: "btn btn-outline-secondary btn-sm ms-auto",
      attrs: { type: "button" },
      text: I18n.get( "projects.own_key_button" ),
      on: { click: () => this.onOwnKey( project ) },
    } );
    return [ join , this.boardLink( project ) , this.githubLink( project ) , extra ].filter( Boolean );
  },

  memberActions( project ) {
    const open = Dom.el( "button" , {
      class: "btn btn-primary btn-sm",
      attrs: { type: "button" , title: I18n.get( "projects.open_hint" ) },
      children: [ Dom.el( "i" , { class: "bi bi-box-arrow-up-right me-1" , attrs: { "aria-hidden": "true" } } ) , Dom.el( "span" , { text: I18n.get( "projects.open_button" ) } ) ],
      on: { click: ( event ) => this.go( event , () => Api.openProject( project.id ) , true ) },
    } );
    const starLabel = I18n.get( project.favorite ? "projects.unstar" : "projects.star" );
    const star = Dom.el( "button" , {
      class: "btn btn-link btn-sm app-icon-button " + ( project.favorite ? "text-warning" : "text-body-secondary" ),
      attrs: { type: "button" , "aria-label": starLabel , title: starLabel , "aria-pressed": String( project.favorite ) },
      children: [ Dom.el( "i" , { class: "bi " + ( project.favorite ? "bi-star-fill" : "bi-star" ) , attrs: { "aria-hidden": "true" } } ) ],
      on: { click: () => this.onStar( project ) },
    } );
    const items = [
      Dom.el( "button" , { attrs: { type: "button" } , text: I18n.get( "projects.new_link_button" ) , on: { click: () => this.onNewLink( project ) } } ),
      Dom.el( "button" , { attrs: { type: "button" } , text: I18n.get( "projects.tags_button" ) , on: { click: () => this.onTags( project ) } } ),
      project.own_key
        ? Dom.el( "button" , { attrs: { type: "button" } , text: I18n.get( "projects.own_key_remove" ) , on: { click: () => this.onRemoveOwnKey( project ) } } )
        : Dom.el( "button" , { attrs: { type: "button" } , text: I18n.get( "projects.own_key_button" ) , on: { click: () => this.onOwnKey( project ) } } ),
      Dom.el( "button" , { class: "btn-outline-danger" , attrs: { type: "button" } , text: I18n.get( "projects.leave_button" ) , on: { click: () => this.onLeave( project ) } } ),
    ];
    const menu = Dom.rowMenu( project.name , items );
    menu.classList.add( "ms-auto" );
    return [ open , this.boardLink( project ) , this.githubLink( project ) , star , menu ].filter( Boolean );
  },

  // The project's TodoBoard board, view only, when an admin gave its link.
  boardLink( project ) {
    if ( !/^https?:\/\//.test( project.board_share_url || "" ) ) return null;
    return Dom.el( "a" , {
      class: "btn btn-outline-secondary btn-sm",
      attrs: { href: project.board_share_url , target: "_blank" , rel: "noopener noreferrer" , title: I18n.get( "projects.board_hint" ) , "aria-label": I18n.get( "projects.board_button" ) },
      children: [ Dom.el( "i" , { class: "bi bi-kanban" , attrs: { "aria-hidden": "true" } } ) ],
    } );
  },

  // Where the project's code lives, when an admin gave its link.
  githubLink( project ) {
    if ( !/^https?:\/\//.test( project.github_url || "" ) ) return null;
    return Dom.el( "a" , {
      class: "btn btn-outline-secondary btn-sm",
      attrs: { href: project.github_url , target: "_blank" , rel: "noopener noreferrer" , title: I18n.get( "projects.github_hint" ) , "aria-label": I18n.get( "projects.github_button" ) },
      children: [ Dom.el( "i" , { class: "bi bi-github" , attrs: { "aria-hidden": "true" } } ) ],
    } );
  },

  // Asks for a link and opens it in a new tab, so the portal stays where it
  // is. The tab has to be opened now, before the await, or the browser takes
  // it for a popup. Join (newTab false) still goes there in this tab.
  async go( event , ask , newTab ) {
    const tab = newTab ? window.open( "about:blank" , "_blank" ) : null;
    if ( newTab && tab ) window.focus();
    try {
      const answer = await ask();
      if ( tab ) {
        tab.opener = null;
        tab.location.href = answer.login_url;
        await this.refresh();
      } else {
        window.location.assign( answer.login_url );
      }
    } catch ( error ) {
      if ( tab ) tab.close();
      App.showError( error.message );
      await this.refresh();
    }
  },

  // The same fresh link Open makes, put on the clipboard instead of opened:
  // for signing in on another device or browser. Safari only lets a page
  // write the clipboard during the click, so the write starts now with the
  // link still to come (a ClipboardItem holding a promise); browsers without
  // that get writeText once it arrives, and failing both the link is shown
  // to copy by hand.
  async onNewLink( project ) {
    const asked = Api.openProject( project.id ).then( ( answer ) => answer.login_url );
    asked.catch( () => {} );
    let copied = false;
    try {
      if ( window.ClipboardItem && navigator.clipboard && navigator.clipboard.write ) {
        const blob = asked.then( ( url ) => new Blob( [ url ] , { type: "text/plain" } ) );
        await navigator.clipboard.write( [ new ClipboardItem( { "text/plain": blob } ) ] );
        copied = true;
      }
    } catch ( error ) { /* falls through to the next way */ }
    let url;
    try {
      url = await asked;
    } catch ( error ) {
      App.showError( error.message );
      await this.refresh();
      return;
    }
    if ( !copied ) {
      try {
        await navigator.clipboard.writeText( url );
        copied = true;
      } catch ( error ) { /* shown below instead */ }
    }
    if ( copied ) Dom.toast( I18n.get( "projects.new_link_copied" ) , "ok" );
    else await this.showLink( project , url );
    await this.refresh();
  },

  showLink( project , url ) {
    return Dom.form( {
      title: I18n.get( "projects.new_link_heading" ).replace( "{name}" , () => project.name ),
      submit: I18n.get( "projects.new_link_done" ),
      fields: [ { name: "link" , type: "custom" , build: ( id ) => {
        const input = Dom.el( "input" , { class: "form-control font-monospace" , attrs: { type: "text" , id: id , readonly: "" , "aria-describedby": id + "-help" } } );
        input.value = url;
        input.addEventListener( "focus" , () => input.select() );
        return { element: Dom.el( "div" , { children: [
          Dom.el( "label" , { class: "form-label" , text: I18n.get( "projects.new_link_label" ) , attrs: { for: id } } ),
          input,
          Dom.el( "div" , { class: "form-text" , text: I18n.get( "projects.new_link_help" ) , attrs: { id: id + "-help" } } ),
        ] } ) , read: () => url };
      } } ],
      save: () => true,
    } );
  },

  find( id ) {
    return this.all.find( ( project ) => project.id === id );
  },

  // Changes one tile in place from the card the server answered.
  replace( card ) {
    const index = this.all.findIndex( ( project ) => project.id === card.id );
    if ( index >= 0 ) this.all[ index ] = card;
    this.fillTags();
    this.render();
  },

  async onStar( project ) {
    const before = project.favorite;
    const set = async ( value ) => this.replace( ( await Api.favoriteProject( project.id , value ) ).project );
    try {
      await Undo.run( {
        label: I18n.get( before ? "projects.undo_unstar" : "projects.undo_star" ),
        do: () => set( !before ),
        undo: () => set( before ),
      } );
    } catch ( error ) { App.showError( error.message ); }
  },

  async onTags( project ) {
    const before = project.tags.slice();
    const set = async ( tags ) => this.replace( ( await Api.tagProject( project.id , tags ) ).project );
    const saved = await Dom.form( {
      title: I18n.get( "projects.tags_heading" ).replace( "{name}" , () => project.name ),
      fields: [ { name: "tags" , label: I18n.get( "projects.tags_label" ) , help: I18n.get( "projects.tags_help" ) , value: before.join( ", " ) , maxLength: 330 } ],
      save: ( values ) => set( values.tags.split( "," ).map( ( tag ) => tag.trim() ).filter( Boolean ) ),
    } );
    if ( saved === null ) return;
    const after = this.find( project.id ).tags.slice();
    Undo.record( { label: I18n.get( "projects.undo_tags" ) , undo: () => set( before ) , redo: () => set( after ) } );
  },

  async onLeave( project ) {
    try {
      await Undo.run( {
        label: I18n.get( "projects.undo_leave" ).replace( "{name}" , () => project.name ),
        toast: true,
        do: async () => {
          await Api.leaveProject( project.id );
          await this.refresh();
        },
        undo: async () => this.replace( ( await Api.restoreProjectMembership( project.id ) ).project ),
      } );
    } catch ( error ) { App.showError( error.message ); }
  },

  // Your own key on that project: the way in for an account the portal's key
  // cannot make links for (the project's owner, its admins), or one you had
  // before the portal. The key is sent once and never shown again.
  async onOwnKey( project ) {
    const saved = await Dom.form( {
      title: I18n.get( "projects.own_key_heading" ).replace( "{name}" , () => project.name ),
      fields: [ {
        name: "key" , type: "password" , required: true , maxLength: 400 ,
        label: I18n.get( "projects.own_key_label" ) , help: I18n.get( "projects.own_key_help" ),
      } ],
      save: ( values ) => Api.setOwnProjectKey( project.id , values.key ),
    } );
    if ( saved === null ) return;
    Dom.toast( I18n.get( "projects.own_key_saved" ).replace( "{name}" , () => saved.remote_name ) , "ok" );
    await this.refresh();
  },

  async onRemoveOwnKey( project ) {
    if ( !( await Dom.confirm( I18n.get( "projects.own_key_remove_confirm" ) ) ) ) return;
    try {
      await Api.removeOwnProjectKey( project.id );
      await this.refresh();
    } catch ( error ) { App.showError( error.message ); }
  },
};
