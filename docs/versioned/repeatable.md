---
title: Repeatable Migrations
id: repeatable
slug: /versioned/repeatable
---

# Repeatable Migrations

Repeatable migrations are special migration files that can be re-applied whenever their content changes. They are typically used for objects like views, stored procedures, or reference data that may need to be updated over time.

## Introduction

While versioned migrations in Atlas run only once in a specific order, repeatable migrations can run multiple times. Atlas tracks repeatable migrations by their name and a checksum of their content. When the content changes, Atlas will re-apply the migration during the next `migrate apply` operation.

Repeatable migrations are particularly useful for:

1. **Database views** that need to be recreated when underlying tables change
2. **Stored procedures and functions** that may need updates without affecting data
3. **Reference data** like lookup tables that need consistent updates
4. **User permissions** that need to be kept in sync

## Creating Repeatable Migrations

To create a repeatable migration, name your file with the `R__` prefix (R for Repeatable) followed by a description. For example:
