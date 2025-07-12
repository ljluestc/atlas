// Apply returns the migrate apply command.
func Apply(cmd *cobra.Command) *cobra.Command {
	var (
		flags struct {
			URL                string
			DirURL             string
			DryRun             bool
			TxMode             string
			Baseline           string
			AllowDirty         bool
			RevisionsSchema    string
			BaselineTableName  string
			Verbose            bool
			DevURL             string
			IncludeSystemTable bool
			ToVersion          string
			Amount             int
			Lock               lockMode
			SkipRepeatable     bool
			RepeatableOnly     bool
		}
		applyCmd = &cobra.Command{
			Use:   "apply",
			Short: "Apply pending migration files on the database",
			Long: `Apply pending migration files on the database.

The 'migrate apply' command brings the database up to date by executing migration files in sequence, 
in an atomic and safe manner. By default, Atlas will execute all pending files to the database. However, 
using the "--amount" flag will limit the number of migration files to execute, and using the "--to-version" 
flag will execute all migrations files up to a specific version. 

On each apply, the 'atlas_schema_revisions' table is updated with information about the applied migrations.

By default, the migration process will be executed atomically using a wrap-all-in-single-tx strategy. 
This strategy can be controlled using the '--tx-mode' flag.

Atlas supports two types of migrations:
1. Versioned migrations (prefixed with a version number) - run once in order
2. Repeatable migrations (prefixed with 'R__') - run whenever their content changes
`,
			Example: examples + `
  # Apply only repeatable migrations
  atlas migrate apply --dir file://migrations --url "mysql://root:pass@localhost:3306/db" --repeatable-only

  # Apply versioned migrations but skip repeatable migrations
  atlas migrate apply --dir file://migrations --url "mysql://root:pass@localhost:3306/db" --skip-repeatable
`,
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				// Flags which are unsupported in dir-based URL format.
				if flags.DirURL != "" && flags.Verbose {
					return errors.New("--verbose is not supported with --dir flag")
				}
				// Ensure the directory reference is well formatted.
				if flags.DirURL != "" {
					// Parse migrate dir.
					u, err := url.Parse(flags.DirURL)
					if err != nil {
						return err
					}
					switch {
					case u.Scheme != "file" && u.Scheme != "atlas":
						return fmt.Errorf("unknown scheme %q", u.Scheme)
					case u.Host != "":
						return errors.New("dir url should not have host part")
					}
				}
				// Cannot use both --skip-repeatable and --repeatable-only
				if flags.SkipRepeatable && flags.RepeatableOnly {
					return errors.New("cannot use both --skip-repeatable and --repeatable-only")
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, _ []string) error {
				return runApply(cmd.Context(), cmd.Flag("dir").Value.String(), cmd.Flag("url").Value.String(), cmdapi.NewLockOptions(cmd, flags.Lock), &flags)
			},
		}
	)
	applyCmd.Flags().SortFlags = false
	applyCmd.Flags().StringVar(&flags.URL, "url", "", "URL to the database using the format:\n\"" + clientpool.DriverFormats() + "\"")
	applyCmd.Flags().StringVar(&flags.DirURL, "dir", "", "URL to migration directory using the format: \n\""+migratedir.FormatWithDefaults+"\"")
	cmdapi.AddFlagLockOptions(applyCmd, &flags.Lock)
	applyCmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "print SQL without executing it")
	applyCmd.Flags().StringVar(&flags.TxMode, "tx-mode", "all", txModeAll.Help())
	applyCmd.Flags().StringVar(&flags.Baseline, "baseline", "", `set a baseline version to allow migrate to run on existing databases.
See: https://atlasgo.io/versioned/apply#existing-database`)
	applyCmd.Flags().BoolVar(&flags.AllowDirty, "allow-dirty", false, "allow start working on a non-clean database")
	applyCmd.Flags().StringVar(&flags.RevisionsSchema, "revisions-schema", "", "database schema where the revisions table resides")
	applyCmd.Flags().StringVar(&flags.BaselineTableName, "baseline-table-name", "", "table name to store baseline information")
	applyCmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "verbose output of SQL statements (for validation purposes)")
	applyCmd.Flags().StringVar(&flags.DevURL, "dev-url", "", fmt.Sprintf(
		"URL for the dev database using the format:\n%q", clientpool.DriverFormats(),
	))
	applyCmd.Flags().BoolVar(&flags.IncludeSystemTable, "include-system-tables", false, "include system tables in analysis")
	applyCmd.Flags().StringVar(&flags.ToVersion, "to-version", "", "apply migrations up to (including) version")
	applyCmd.Flags().IntVar(&flags.Amount, "amount", 0, "limit the number of migrations to apply")
	applyCmd.Flags().BoolVar(&flags.SkipRepeatable, "skip-repeatable", false, "skip repeatable migrations")
	applyCmd.Flags().BoolVar(&flags.RepeatableOnly, "repeatable-only", false, "apply only repeatable migrations")
	applyCmd.MarkFlagsMutuallyExclusive("amount", "to-version")
	applyCmd.MarkFlagsMutuallyExclusive("skip-repeatable", "repeatable-only")

	applyCmd.MarkFlagsMutuallyExclusive("url", "dev-url")
	applyCmd.Flags().MarkHidden("dev-url")
	applyCmd.Flags().MarkHidden("include-system-tables")

	// For backward compatibility.
	applyCmd.Flags().MarkDeprecated("verbose", "use --dry-run instead")

	return applyCmd
}

// runApply updated to handle repeatable migrations
func runApply(ctx context.Context, dirURL, dbURL string, lock *cmdapi.LockOptions, flags *struct {
	URL                string
	DirURL             string
	DryRun             bool
	TxMode             string
	Baseline           string
	AllowDirty         bool
	RevisionsSchema    string
	BaselineTableName  string
	Verbose            bool
	DevURL             string
	IncludeSystemTable bool
	ToVersion          string
	Amount             int
	Lock               lockMode
	SkipRepeatable     bool
	RepeatableOnly     bool
}) error {
	// Existing runApply code...
	
	// Add flag handling for repeatable migrations
	// For example:
	options := []migrate.ApplyOption{
		migrate.WithSkipRepeatable(flags.SkipRepeatable),
		migrate.WithRepeatableOnly(flags.RepeatableOnly),
	}
	
	// Pass options to the migrate.Apply function
	// ...
	
	return nil
}
