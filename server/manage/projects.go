package manage

import (
	bufio "bufio"
	flag "flag"
	fmt "fmt"
	io "io"
	os "os"

	config "projectportal/server/config"
	control "projectportal/server/control"
)

// `manage list-projects` and `manage set-project-key`: the catalog from the
// shell. Neither is on the backend interface, which is the template's; they
// take whichever way in the command found (the control socket, or the
// database when the server is stopped).

func commandListProjects( back backend ) ( err error ) {
	var rows []control.ProjectResponse
	switch b := back.( type ) {
	case *controlBackend:
		rows , err = b.client.ListProjects()
	case *directBackend:
		rows , err = control.ListProjectRows( b.store )
	default:
		err = fmt.Errorf( "list-projects: no way in" )
	}
	if err != nil { return }
	for _ , row := range rows {
		key := "no key"
		if row.HasKey { key = "key " + row.KeyStatus }
		fmt.Printf( "%-4d %-32s %-48s %s\n" , row.ID , row.Name , row.BaseURL , key )
	}
	return
}

// commandSetProjectKey reads the key from stdin, so it can be piped from
// another app's `manage portal-connect` without ever being shown.
func commandSetProjectKey( back backend , cfg *config.Config , arguments []string ) ( err error ) {
	set := flag.NewFlagSet( "set-project-key" , flag.ExitOnError )
	which := set.String( "project" , "" , "the project's id, name or address (manage list-projects)" )
	set.Parse( arguments )
	if *which == "" {
		err = fmt.Errorf( "-project is required; the key is read from stdin" )
		return
	}
	line , read_err := bufio.NewReader( io.LimitReader( os.Stdin , 4096 ) ).ReadString( '\n' )
	if read_err != nil && read_err != io.EOF {
		err = read_err
		return
	}
	var row control.ProjectResponse
	switch b := back.( type ) {
	case *controlBackend:
		row , err = b.client.SetProjectKey( *which , line )
	case *directBackend:
		row , err = control.SetProjectKey( b.store , cfg.SecretKey , *which , line )
	default:
		err = fmt.Errorf( "set-project-key: no way in" )
	}
	if err != nil { return }
	fmt.Printf( "admin key set for #%d %s (%s)\n" , row.ID , row.Name , row.BaseURL )
	return
}
