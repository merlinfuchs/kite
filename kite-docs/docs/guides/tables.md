---
sidebar_position: 5
---

# Using Tables

Tables let your Discord bot store information in a database-like format.
Each table has columns, and each row contains one record. For example, you
could use a table to store warnings, support tickets, points, or user
settings.

This guide explains how to create a table, edit its data, use it in flows,
and recover from mistakes.

## What you will build

In this tutorial, we will create a `warnings` table with these columns:

| Column | Type | Purpose |
| --- | --- | --- |
| `user_id` | Text | The Discord user who received the warning |
| `reason` | Text | Why the warning was created |
| `count` | Number | The warning number |
| `active` | Boolean | Whether the warning is still active |
| `created_at` | Date | When the warning was created |

The table will look like this:

| ID | User ID | Reason | Count | Active | Created at |
| --- | --- | --- | --- | --- | --- |
| Generated automatically | `123456789` | Spam | `1` | `true` | A date and time |

## Before you start

You need:

1. A Kite account.
2. An app in Kite.
3. Permission to edit that app.

Tables are separate from stored variables. Existing stored variables continue
to work, but new database-style data should usually go into a table.

## 1. Create a table

1. Open your app in Kite.
2. Open **Tables** in the app navigation.
3. Click **Create table**.
4. Enter `warnings` as the table name.
5. Add the columns from the example above.
6. Click **Create table**.

Column names become column IDs. Use short, clear names with letters, numbers,
and underscores. Names such as `user_id` and `created_at` are easier to use in
flows than names such as `The user's Discord ID`.

Every row receives an ID automatically. You do not need to create or manage
row IDs yourself.

## 2. Choose the correct column type

Kite supports these column types:

- **Text**: Words, sentences, usernames, Discord IDs, and other strings.
- **Number**: Whole numbers or decimal numbers.
- **Boolean**: `true` or `false`.
- **Date**: A date and time. Dates are stored using an ISO/RFC 3339 format.
- **JSON**: Structured data containing objects or arrays.
- **Reference**: A row ID from another table.

Choose the type based on how you will use the value. For example, a warning
count should be a **Number**, not Text. This makes comparisons and arithmetic
in flows reliable.

### Required columns

Mark a column as **Required** when every row must contain a value. A row
cannot be saved while a required column is empty.

Do not mark a column required unless your bot can always provide it. For
example, `reason` might be required, while an optional moderator note might
not be.

### References between tables

Use a **Reference** column when one table points to another table. For
example, a `tickets` table could have a `customer_id` reference to a
`customers` table.

References store the target row's generated ID. They do not copy the target
row's contents. If you need the target's name or email, read the referenced
row in a separate flow.

## 3. Add and edit rows

Open a table to see its data.

### Add a row

1. Click **Add row**.
2. Enter values in the new row.
3. Leave the cell to save the value.

The row ID is generated automatically.

### Edit a row

Click a cell and change its value. Text, number, date, and JSON cells save
when you leave the editor. Boolean cells use a custom menu with **Empty**,
**True**, and **False**.

JSON cells must contain valid JSON. For example:

```json
{"source":"moderator","severity":2}
```

### Sort and select rows

Click a column heading to sort it. Click it again to reverse the order.
Click it a third time to remove sorting.

Use the checkboxes to select rows. The bulk delete action removes all selected
rows after you confirm it. Deletions are permanent unless the operation is
available in the table's history and can be undone.

## 4. Change a table's schema

Click **Edit schema** on the table page.

You can:

- Rename the table.
- Rename columns.
- Change column types.
- Mark columns as required or optional.
- Set or remove a reference target.
- Add columns.
- Remove columns.

Click **Save schema** when you are finished.

### Be careful when changing columns

Changing a column type can make existing values invalid. Before making a
large schema change:

1. Open the table's **History** view.
2. Confirm that the current data is correct.
3. Make one schema change at a time.
4. Check several rows after saving.

Removing a column can make the values in that column inaccessible. Export or
copy important data before removing it.

## 5. Use table blocks in a flow

Tables can be controlled by flow blocks. In the flow editor, search for the
**Tables** category.

The available blocks are:

- **Insert table row**: Creates a new row.
- **Get table row**: Gets one row by its ID.
- **Find table rows**: Finds rows matching filters.
- **Update table row**: Replaces a row's data.
- **Delete table row**: Deletes one row.
- **List table columns**: Gets a table's column definitions.
- **Add table column**: Adds a column.
- **Update table column**: Changes a column.
- **Delete table column**: Removes a column.

### Insert a row

Use **Insert table row** and provide:

- The table ID.
- A JSON object containing the row data.

Example:

```json
{
  "user_id": "123456789",
  "reason": "Spam",
  "count": 1,
  "active": true
}
```

The block returns the created row, including its generated `id`.

### Get a row

Use **Get table row** with:

- The table ID.
- The row ID.

If the row does not exist, the block returns an empty/null result. Check the
result before using its fields.

### Find rows

Use **Find table rows** with a JSON filter array. Each filter contains a
column, an operator, and a value.

Example:

```json
[
  {
    "column": "user_id",
    "operator": "equals",
    "value": "123456789"
  },
  {
    "column": "active",
    "operator": "equals",
    "value": true
  }
]
```

Supported operators include:

- `equals` or `eq`
- `not_equals` or `neq`
- `contains` for text
- `greater_than` or `gt` for numbers
- `less_than` or `lt` for numbers

Set a limit between `1` and `100`. Keeping the limit low is safer and faster.

### Update a row

Use **Update table row** with the table ID, row ID, and a JSON object. The
provided object becomes the row's data, so include all values you want to
keep.

### Delete a row

Use **Delete table row** with the table ID and row ID. Add conditions before
this block when the target comes from user input. This prevents accidentally
deleting the wrong record.

## 6. Use results from table blocks

Table block results can be used by later blocks through the result picker.
For a row, common fields include:

- `id`
- Your column IDs, such as `user_id`
- Your column values, such as `reason` or `count`

For a find operation, the result is an array of rows. Use a loop or an array
operation when you need to process every matching row.

## 7. View history and undo changes

Open a table and click **History** to see recent changes.

History records:

- Table creation and updates.
- Row creation, updates, and deletions.
- Changes made through table flow blocks.

Each entry shows the action, target row when applicable, and time.

To undo an entry:

1. Open **History**.
2. Find the change.
3. Click **Undo**.
4. Confirm the action.

Undo restores the previous snapshot and creates a new history entry. An
already-undone entry cannot be undone again. This prevents the same action
from being reversed repeatedly.

History is not a replacement for backups. For important data, keep an
external export or recreate the data from its original source.

## 8. A complete warning flow

A basic warning command can work like this:

1. Add a command or event listener that receives a user and reason.
2. Use **Find table rows** on `warnings`.
3. Filter by `user_id` and `active = true`.
4. Count the matching rows or calculate the next warning number.
5. Use **Insert table row** with the user ID, reason, count, active status, and
   current date.
6. Send a confirmation message.

Before using **Insert table row**, validate that the reason is not empty and
that the target user is valid. Before using **Delete table row**, check that
the row ID exists and that the command invoker has permission.

## Common problems

### “Column is required”

The row is missing a value in a required column. Enter a value or make the
column optional in **Edit schema**.

### “Invalid row”

One of the values does not match the column type. Check that:

- Numbers are actual numbers, not quoted text.
- Booleans are `true` or `false`.
- Dates use a valid date/time format.
- JSON is valid JSON.

### A reference cannot be selected

The reference target must be another table. A table cannot reference itself
through the schema editor.

### A flow finds no rows

Check the column ID, operator, and value. `user_id` is different from
`username`, and number comparisons do not work correctly when numbers are
stored as text.

### Undo is unavailable

Some history entries are informational or have already been undone. Only
entries with a reversible data snapshot can be undone.

### The table is not visible

Make sure you are in the correct app and that your account has access to it.
Refresh the Tables page after creating or deleting a table.

## Recommended practices

- Use one table for one kind of record.
- Use stable, descriptive column IDs.
- Keep table and column names short.
- Use the correct data type from the beginning.
- Mark only genuinely necessary columns as required.
- Limit flow queries to the smallest useful number of rows.
- Validate user-provided IDs before update or delete operations.
- Use confirmation steps for destructive flows.
- Review History after large edits.
- Keep backups for important or irreplaceable data.

