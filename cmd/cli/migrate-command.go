package main

// doMigrate build the migrate command to running up and down migration to the database
func doMigrate(arg3, arg4 string) error {
	dsn, err := getDSN()
	if err != nil {
		return err
	}

	switch arg3 {
	case "up":
		err := sauri2.UpMigrate(dsn)
		if err != nil {
			return err
		}
	case "down":
		// empty the entire database
		if arg4 == "all" {
			err := sauri2.DownMigrate(dsn)
			if err != nil {
				return err
			}
		} else {
			// drop the most current added migration
			err := sauri2.StepsMigrate(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		// pull down all the migrations added and re-add them again to the database
		err := sauri2.DownMigrate(dsn)
		if err != nil {
			return err
		}
		err = sauri2.UpMigrate(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}
	return nil
}
