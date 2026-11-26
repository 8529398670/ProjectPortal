// One theme everywhere: the theme you use here is sent to every app you are
// on (POST /api/account/theme/send, server/routes/theme_share.go). While the
// switch is on (the default) it goes whenever the theme in use changes:
// another pick, an edit, light/dark. Send to my apps now does it by hand and
// lists what each app said. The theme itself goes from here, because the
// built-in ones live in theme.js and nowhere on the server.
const ThemeShare = {
  on: true,
  last: "",
  timer: null,

  // The fields of the Theme schema, so nothing the browser adds goes along.
  fields: [ "id" , "name" , "mode" , "accent" , "second" , "background" , "surface" , "overrides" , "button" , "light" , "dark" ],

  init( me ) {
    this.on = me.theme_share !== false;
    const toggle = Dom.get( "theme-share-on" );
    toggle.checked = this.on;
    toggle.addEventListener( "change" , () => this.onToggle( toggle ) );
    Dom.get( "theme-share-send" ).addEventListener( "click" , () => this.sendNow() );
    // What the account already uses was sent before, or is the app's own:
    // only a change from here on is sent.
    this.last = this.signature();
    Theme.onChange.push( () => this.changed() );
  },

  payload() {
    const theme = Theme.named( Theme.find( Theme.current ) );
    const out = {};
    this.fields.forEach( ( field ) => { if ( theme[ field ] !== undefined && theme[ field ] !== "" ) out[ field ] = theme[ field ]; } );
    return out;
  },

  signature() {
    return JSON.stringify( this.payload() );
  },

  // A few quick changes (picks, saves) send once, a moment after the last.
  changed() {
    if ( !this.on || this.signature() === this.last ) return;
    clearTimeout( this.timer );
    this.timer = setTimeout( () => this.send( false ) , 1500 );
  },

  async onToggle( toggle ) {
    try {
      const reply = await Api.setThemeShare( toggle.checked );
      this.on = reply.theme_share;
      if ( this.on ) this.changed();
    } catch ( error ) {
      toggle.checked = this.on;
      App.showError( error.message );
    }
  },

  async sendNow() {
    const button = Dom.get( "theme-share-send" );
    button.disabled = true;
    try {
      await this.send( true );
    } finally {
      button.disabled = false;
    }
  },

  // by_hand shows each app's answer; a send on change stays quiet unless an
  // app refused.
  async send( by_hand ) {
    clearTimeout( this.timer );
    const signature = this.signature();
    let reply;
    try {
      reply = await Api.sendTheme( this.payload() );
    } catch ( error ) {
      if ( by_hand ) App.showError( error.message );
      return;
    }
    this.last = signature;
    const sent = reply.apps.filter( ( app ) => app.status === "sent" ).length;
    const refused = reply.apps.some( ( app ) => app.status === "refused" );
    if ( by_hand || refused ) this.show( reply.apps );
    if ( by_hand && reply.apps.length ) Dom.toast( I18n.get( "theme_share.sent_toast" ).replace( "{count}" , () => String( sent ) ) , "ok" );
  },

  show( apps ) {
    const list = Dom.get( "theme-share-results" );
    const tones = { sent: "success" , skipped: "secondary" , refused: "danger" };
    const rows = apps.length ? apps.map( ( app ) => Dom.el( "li" , { class: "list-group-item px-0 d-flex flex-wrap align-items-center gap-2" , children: [
      Dom.badge( I18n.get( "theme_share.status_" + app.status ) , tones[ app.status ] || "secondary" ),
      Dom.el( "span" , { class: "fw-semibold" , text: app.name } ),
      app.note ? Dom.el( "span" , { class: "small text-body-secondary text-break" , text: app.note } ) : null,
    ].filter( Boolean ) } ) ) : [ Dom.el( "li" , { class: "list-group-item px-0 text-body-secondary" , text: I18n.get( "theme_share.none" ) } ) ];
    Dom.swap( list , ...rows );
    Dom.show( list , true );
  },
};
