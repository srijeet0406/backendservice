package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/srijeet0406/backendservice/config"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	DBSuperUser = `postgres`
	// available commands
	CmdCreateDB        = "createdb"
	CmdDropDB          = "dropdb"
	CmdCreateUser      = "create_user"
	CmdDropUser        = "drop_user"
	CmdShowUsers       = "show_users"
	CmdLoadSchema = "load_schema"
	// defaults
	defaultDBDir            = "./"
	defaultDBConfigPath     = defaultDBDir + "dbconf.conf"
	defaultDBMigrationsPath = defaultDBDir + "migrations"
	defaultDBSeedsPath      = defaultDBDir + "seeds.sql"
	defaultDBSchemaPath     = defaultDBDir + "create_tables.sql"
	defaultDBPatchesPath    = defaultDBDir + "patches.sql"
	defaultEnvironment = "development"
)

var (
	dbConfigPath    = defaultDBConfigPath
	dbMigrationsDir = defaultDBMigrationsPath
	dbSeedsPath     = defaultDBSeedsPath
	dbSchemaPath    = defaultDBSchemaPath
	dbPatchesPath   = defaultDBPatchesPath
)

var (
	// globals that are passed in via CLI flags and used in commands
	Environment    string
)

func die(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func createDB(cfg config.DBConfig) {
	dbExistsCmd := exec.Command("psql", "-h", cfg.Hostname, "-U", DBSuperUser, "-p", cfg.Port, "-tAc", "SELECT 1 FROM pg_database WHERE datname='"+cfg.DBName+"'")
	stderr := bytes.Buffer{}
	dbExistsCmd.Stderr = &stderr
	out, err := dbExistsCmd.Output()
	// An error is returned if the database could not be found, which is to be expected. Don't exit on this error.
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to check if DB already exists: "+err.Error()+", stderr: "+stderr.String())
	}
	if len(out) > 0 {
		fmt.Println("Database " + cfg.DBName + " already exists")
		return
	}
	createDBCmd := exec.Command("createdb", "-h", cfg.Hostname, "-p", cfg.Port, "-U", DBSuperUser, "-e", "--owner", cfg.User, cfg.DBName)
	out, err = createDBCmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't create db " + cfg.DBName + ": " + err.Error())
	}
}

func dropDB(cfg config.DBConfig) {
	fmt.Println("Dropping database: " + cfg.DBName)
	cmd := exec.Command("dropdb", "-h", cfg.Hostname, "-p", cfg.Port, "-U", DBSuperUser, "-e", "--if-exists", cfg.DBName)
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't drop db " + cfg.DBName + ": " + err.Error())
	}
}

func createUser(cfg config.DBConfig) {
	fmt.Println("Creating user: " + cfg.User)
	userExistsCmd := exec.Command("psql", "-h", cfg.Hostname, "-U", DBSuperUser, "-p", cfg.Port, "-tAc", "SELECT 1 FROM pg_roles WHERE rolname='"+cfg.User+"'")
	stderr := bytes.Buffer{}
	userExistsCmd.Stderr = &stderr
	out, err := userExistsCmd.Output()
	// An error is returned if the user could not be found, which is to be expected. Don't exit on this error.
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to check if user already exists: "+err.Error()+", stderr: "+stderr.String())
	}
	if len(out) > 0 {
		fmt.Println("User " + cfg.User + " already exists")
		return
	}
	createUserCmd := exec.Command("psql", "-h", cfg.Hostname, "-p", cfg.Port, "-U", DBSuperUser, "-etAc", "CREATE USER "+cfg.User+" WITH LOGIN ENCRYPTED PASSWORD '"+cfg.Password+"'")
	out, err = createUserCmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't create user " + cfg.User)
	}
}

func dropUser(cfg config.DBConfig) {
	cmd := exec.Command("dropuser", "-h", cfg.Hostname, "-p", cfg.Port, "-U", DBSuperUser, "-i", "-e", cfg.User)
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't drop user " + cfg.User)
	}
}

func showUsers(cfg config.DBConfig) {
	cmd := exec.Command("psql", "-h", cfg.Hostname, "-p", cfg.Port, "-U", DBSuperUser, "-ec", `\du`)
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't show users")
	}
}

func main() {
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usage()) }

	var shortCfg string
	var longCfg string
	flag.StringVar(&shortCfg, "c", "", "Provide a path to a database configuration file, instead of using the default (./db/dbconf.yml or ./db/trafficvault/dbconf.yml for Traffic Vault)")
	flag.StringVar(&longCfg, "config", "", "Provide a path to a database configuration file, instead of using the default (./db/dbconf.yml or ./db/trafficvault/dbconf.yml for Traffic Vault)")

	var shortEnv string
	var longEnv string
	flag.StringVar(&shortEnv, "e", "", "Use configuration for environment ENV (defined in the database configuration file)")
	flag.StringVar(&longEnv, "env", "", "Use configuration for environment ENV (defined in the database configuration file)")


	flag.Parse()
	collapse(shortCfg, longCfg, "config", defaultDBConfigPath, &dbConfigPath)
	collapse(shortEnv, longEnv, "environment", defaultEnvironment, &Environment)

	fmt.Println(flag.Args())
	if len(flag.Args()) != 1 || flag.Arg(0) == "" {
		die(usage())
	}
	if Environment == "" {
		die(usage())
	}
	commands := make(map[string]func(dbConfig config.DBConfig))

	dbCfg, err := ParseDBConfig(dbConfigPath)
	if err != nil {
		die(err.Error())
	}
	if dbCfg == nil {
		die("couldn't parse db config from conf file")
	}
	commands[CmdCreateDB] = createDB
	commands[CmdDropDB] = dropDB
	commands[CmdCreateUser] = createUser
	commands[CmdDropUser] = dropUser
	commands[CmdShowUsers] = showUsers
	commands[CmdLoadSchema] = loadSchema

	userCmd := flag.Arg(0)
	if cmd, ok := commands[userCmd]; ok {
		cmd(*dbCfg)
	} else {
		die("invalid command: " + userCmd + "\n" + usage())
	}
}

func loadSchema(cfg config.DBConfig) {
	fmt.Println("Creating database tables.")
	schemaPath := dbSchemaPath
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		die("unable to read '" + schemaPath + "': " + err.Error())
	}
	cmd := exec.Command("psql", "-h", cfg.Hostname, "-p", cfg.Port, "-d", cfg.DBName, "-U", cfg.User, "-e", "-v", "ON_ERROR_STOP=1")
	cmd.Stdin = bytes.NewBuffer(schemaBytes)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		die("Can't create database tables")
	}
}

func ParseDBConfig(path string) (*config.DBConfig, error) {
	var dbConfig config.DBConfig
	confBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading DB conf '%s': %w", path, err)
	}

	err = json.Unmarshal(confBytes, &dbConfig)
	if err != nil {
		return nil, errors.New("unmarshalling DB conf yaml: " + err.Error())
	}
	return &dbConfig, nil
}

func collapse(o1, o2, name, def string, dest *string) {
	if o1 == "" {
		if o2 == "" {
			*dest = def
			return
		}
		*dest = o2
		return
	}
	if o2 == "" {
		*dest = o1
		return
	}
	if o1 != o2 {
		die("conflicting definitions of '" + name + "' - must be specified only once\n" + usage())
	}
	*dest = o1
	return
}

func usage() string {
	programName := os.Args[0]
	var buff strings.Builder
	buff.WriteString("Usage: ")
	buff.WriteString(programName)
	buff.WriteString(` [OPTION] OPERATION [ARGUMENT(S)]

-c, --config CFG         Provide a path to a database configuration file,
                         instead of using the default (./db/dbconf.yml or
                         ./db/trafficvault/dbconf.yml for Traffic Vault)
-e, --env ENV            Use configuration for environment ENV (defined in
                         the database configuration file)
-h, --help               Show usage information and exit
-m, --migrations-dir DIR Use DIR as the migrations directory, instead of the
                         default (./db/migrations/ or
                         ./db/trafficvault/migrations for Traffic Vault)
-p, --patches PATCH      Provide a path to a set of database patch statements,
                         instead of using the default (./db/patches.sql)
-s, --schema SCHEMA      Provide a path to a schema file, instead of using the
                         default (./db/create_tables.sql or
                         ./db/trafficvault/create_tables.sql for Traffic Vault)
-S, --seeds SEEDS        Provide a path to a seeds statements file, instead of
                         using the default (./db/seeds.sql)
-v, --trafficvault       Perform operations for Traffic Vault instead of the
                         Traffic Ops database

OPERATION      The operation to perform; one of the following:

    migrate     - Execute migrate (without seeds or patches) on the database for the
                  current environment.
    up          - Alias for 'migrate'
    down        - Roll back a single migration from the current version.
    createdb    - Execute db 'createdb' the database for the current environment.
    dropdb      - Execute db 'dropdb' on the database for the current environment.
    create_migration NAME
                - Creates a pair of timestamped up/down migrations titled NAME.
    create_user - Execute 'create_user' the user for the current environment
                  (traffic_ops).
    dbversion   - Prints the current migration timestamp
    drop_user   - Execute 'drop_user' the user for the current environment
                  (traffic_ops).
    patch       - Execute sql from db/patches.sql for loading post-migration data
                  patches (NOTE: not supported with --trafficvault option).
    redo        - Roll back the most recently applied migration, then run it again.
    reset       - Execute db 'dropdb', 'createdb', load_schema, migrate on the
                  database for the current environment.
    seed        - Execute sql from db/seeds.sql for loading static data (NOTE: not
                  supported with --trafficvault option).
    show_users  - Execute sql to show all of the user for the current environment.
    status      - Prints the current migration timestamp (Deprecated, status is now an
                  alias for dbversion and will be removed in a future Traffic
                  Control release).
    upgrade     - Execute migrate, seed, and patches on the database for the current
                  environment.`)
	return buff.String()
}