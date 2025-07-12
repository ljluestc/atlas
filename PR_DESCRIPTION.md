# Support Repeatable Migrations (#2155)

## Description
This PR implements support for repeatable migrations in Atlas, as requested in issue #2155. Repeatable migrations are special migration files that can be re-applied whenever their content changes. Similar to how Flyway and Liquibase handle repeatable migrations, this feature allows users to maintain database objects like views, stored procedures, and reference data that may need to be updated over time.

## Implementation Details
The implementation includes:

1. **Repeatable Migration Format**: Added support for a new migration file naming pattern `R__description.sql` that indicates a repeatable migration (R for Repeatable).

2. **Checksum Tracking**: Implemented a mechanism to track applied repeatable migrations with checksums of their content in a new `atlas_repeatable_migrations` table.

3. **Selective Re-application**: Added logic to only re-apply repeatable migrations when their content changes (detected by checksum comparison).

4. **Configuration Options**: Added command-line flags to control repeatable migration behavior:
   - `--skip-repeatable` to skip repeatable migrations
   - `--repeatable-only` to apply only repeatable migrations

5. **Migration Engine Updates**: Enhanced the migration engine to handle both versioned and repeatable migrations with appropriate execution order (versioned first, then repeatable).

6. **Testing**: Added comprehensive tests for repeatable migrations, including various scenarios like applying all migrations, skipping repeatable migrations, applying only repeatable migrations, and handling changed content.

7. **Documentation**: Created detailed documentation for repeatable migrations in `docs/versioned/repeatable.md`.

## Example Usage
Users can create and use repeatable migrations as follows:

1. Create a repeatable migration file:
   ```sql
   -- R__create_user_view.sql
   CREATE OR REPLACE VIEW user_details AS
   SELECT u.id, u.username, p.first_name, p.last_name
   FROM users u
   JOIN profiles p ON u.id = p.user_id;
   ```

2. Apply all migrations (both versioned and repeatable):
   ```bash
   atlas migrate apply --dir file://migrations --url "mysql://root:pass@localhost:3306/db"
   ```

3. Apply only repeatable migrations:
   ```bash
   atlas migrate apply --dir file://migrations --url "mysql://root:pass@localhost:3306/db" --repeatable-only
   ```

## Benefits
This feature addresses several use cases:

1. **Maintainable Views**: Keep database views in sync with changing underlying table structures
2. **Evolving Procedures**: Update stored procedures and functions without versioning overhead
3. **Reference Data**: Maintain consistent reference/lookup data across environments
4. **Permissions**: Keep user permissions in sync with evolving application needs

## How It Works
Atlas implements repeatable migrations with the following process:

1. After applying versioned migrations, Atlas checks for repeatable migrations in the migration directory
2. For each repeatable migration, Atlas calculates a SHA-256 checksum of its content
3. Atlas compares this checksum with the previously applied version (stored in the database)
4. If the migration is new or its content has changed (different checksum), Atlas applies it
5. Unchanged migrations are skipped

Repeatable migrations are applied in alphabetical order by filename, giving users control over execution order through naming conventions.

## Testing
The implementation includes thorough tests in `sql/migrate/migrate_test.go`, covering:
- Detection of repeatable migration files
- Proper application of new repeatable migrations
- Skipping unchanged repeatable migrations
- Re-applying changed repeatable migrations
- Using the `--skip-repeatable` and `--repeatable-only` flags

All tests pass successfully.

## Related Issues
Closes #2155
