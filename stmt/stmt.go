package stmt

const CreateTables = `
CREATE TABLE IF NOT EXISTS txtmsg
(
	id      text   PRIMARY KEY COLLATE NOCASE,
	msg     text   NOT NULL,
	ctime   int   NOT NULL,
	mtime   int   NOT NULL,
);

CREATE INDEX IF NOT EXISTS idx_mima_ctime ON txtmsg(ctime);
CREATE INDEX IF NOT EXISTS idx_mima_mtime ON txtmsg(mtime);

CREATE TABLE IF NOT EXISTS metadata
(
  name         text   NOT NULL UNIQUE,
  int_value    int    NOT NULL DEFAULT 0,
  text_value   text   NOT NULL DEFAULT "" 
);
`

const InsertIntValue = `INSERT INTO metadata (name, int_value) VALUES (?, ?);`
const GetIntValue = `SELECT int_value FROM metadata WHERE name=?;`
const UpdateIntValue = `UPDATE metadata SET int_value=? WHERE name=?;`

const InsertTextValue = `INSERT INTO metadata (name, text_value) VALUES (?, ?);`
const GetTextValue = `SELECT text_value FROM metadata WHERE name=?;`
const UpdateTextValue = `UPDATE metadata SET text_value=? WHERE name=?;`
