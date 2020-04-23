# njson2csv

Tool for converting new-line delimited JSON (NJSON) to CSV format.

# usage

`njson2csv [-o output.csv] njson_file [more_njson_files...]`

# notes/features

* Nested dictionaries (of arbitrary depth) are handled by apending child to parent field names sperated by a `-`:


  `{"parent": {"child": true}, "foo": "bar"}`

  becomes...

  ```
  parent_child,foo,
  true,bar,
  ```

* Multiple NJSON files can be combined into a single csv; if they have different fields, all fields will be added with black cells inserted where no data is present
  `njson2csv file1.log file2.log [...] fileN.log`

* Will always add a csv field `__nsjon_file` with the filename that those fields came from. (No more loosing what belongs to which file!!1!)

* List values are added as their JSON string value to a cell -- the elements are not expanded into their own fields ¯\\_(ツ)_/¯

  `{"some_list": [1, 2, 3]}`

  becomes...

  ```
  some_list,
  [1,2,3],
  ```

* Filter to a subset of columns;

	`$ njson2csv --columns `col1,col2,col4``


Enjoy!
