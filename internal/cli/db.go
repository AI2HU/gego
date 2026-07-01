package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/services"
)

var (
	upgradeSQLitePath  string
	upgradePostgresURI string
	upgradeConfigPath  string
	upgradeDryRun      bool
	upgradeForce       bool
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database maintenance commands",
}

var dbUpgradeFromSQLiteCmd = &cobra.Command{
	Use:   "upgrade-from-sqlite",
	Short: "Migrate data from legacy SQLite to PostgreSQL",
	RunE:  runDBUpgradeFromSQLite,
}

func init() {
	dbUpgradeFromSQLiteCmd.Flags().StringVar(&upgradeSQLitePath, "sqlite-path", "", "Source SQLite database path")
	dbUpgradeFromSQLiteCmd.Flags().StringVar(&upgradePostgresURI, "postgres-uri", "", "Target PostgreSQL URI")
	dbUpgradeFromSQLiteCmd.Flags().StringVar(&upgradeConfigPath, "config", "", "Config file to update after migration")
	dbUpgradeFromSQLiteCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "Print row counts without migrating")
	dbUpgradeFromSQLiteCmd.Flags().BoolVar(&upgradeForce, "force", false, "Allow migrating into a non-empty PostgreSQL database")
	dbCmd.AddCommand(dbUpgradeFromSQLiteCmd)
}

func runDBUpgradeFromSQLite(cmd *cobra.Command, args []string) error {
	configPath := upgradeConfigPath
	if configPath == "" {
		if cfgFile != "" {
			configPath = cfgFile
		} else {
			configPath = config.GetConfigPath()
		}
	}

	appCfg := cfg
	if appCfg == nil {
		var err error
		appCfg, err = config.ResolveConfig(configPath, true)
		if err != nil {
			return err
		}
	}

	svc := services.NewUpgradeService(appCfg)
	result, err := svc.Run(context.Background(), services.UpgradeSQLiteToPostgres, services.UpgradeRunOptions{
		SQLitePath:  upgradeSQLitePath,
		PostgresURI: upgradePostgresURI,
		ConfigPath:  configPath,
		DryRun:      upgradeDryRun,
		Force:       upgradeForce,
	})
	if err != nil {
		return err
	}

	fmt.Printf("✅ Upgrade %s\n", result.Status)
	fmt.Println(result.Message)
	if result.RestartRequired {
		fmt.Println("\nRestart the API and worker to use PostgreSQL.")
	}
	return nil
}
