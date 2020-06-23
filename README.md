# LookML Generator

[Looker](https://looker.com/) has a nice code generator from existing tables. But when columns are added there is no feature on looker to add missing columns to the view.

This tool is a very early work in progress to render LookML for new columns and optionally show columns implemented in LookML that do not exist in the database.

This is a quick and dirty implementation using Redshift output.


## Usage 

1. Dump table description to CSV file
2. Pull latest LookML from GitHub
3. Run app to generate mission LookML

**Command Line Arguments**

|Name | | |
|---|---|---|
|check | |Will report any columns found in LookML that are not in the table |
|lookml | |The LookML view that represents the table |
|suffix | |Sometimes columns have an annoying suffix like `_c`. If the suffix is passed in that suffix is removed from the dimension. |
|table | |The CSV that contains the table description |
|verbose | |Prints debug information to console |


**Example:**

`looker_new_column_helper -table=./input.csv -lookml=accounts.view.lkml -suffix="_c" -check=true`

### Dump Table Description To CSV

In looker you can run the following SQL command or go to table in Looker and select "Describe" button. Once Looker generates results save as CSV

```
SELECT column_name,
       data_type,
       character_maximum_length
FROM   information_schema.columns
WHERE  table_schema = '{{SCHEMA}}'
       AND table_name = '{{TABLE}}'
ORDER BY column_name ASC

```

The following fields need replaced:

* `{{SCHEMA}}` with the schema name
* `{{TABLE}}` with the table name


