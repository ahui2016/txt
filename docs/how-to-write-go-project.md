# How to Write a Go Project (从零开始写一个 Go 语言小工具)

今天有灵感做一个小工具，然后发现这个工具很简单，但也包含了前端、后端的一些基本操作，非常适合用来介绍一个小工具的制作过程。

## 需求与灵感

小工具灵感通常起源于自己的一个小需求。

现在我有一个需求：需要在不同的电脑、手机之间交换纯文本（一个短字符串），包括 Windows, MacOS, Linux, iOS, Andriod, 其中包括无图形界面的服务器。

做一个满足这个需求的小工具，对于我来说最简单的就是做一个网站，由于我的需求只是传输很短的纯文本，并且是一个自用的工具，因此不需要担心传输速度、流量、服务器负荷等问题。

## 技术选型

由于我很懒，也怕麻烦，技术实力也差，因此我会尽量选择简单、原始、直观（思维负担低）的技术栈。

比如前端我就用 JQuery, 简单到极致。我也用过 React 和 Vue, 结果发现对于一个简单的界面来说, JQuery/React/Vue 都差不多！

React/Vue 做复杂界面有优势，但对于简单界面，实在没有带来特别明显的好处。JQuery 的好处是比较轻，而且可以彻底抛弃 npm, 这使得前端构建省了很多事。

当然，这是因为我不太爱玩前端，对于爱玩前端的人来说选择就完全不一样了，这需要根据自己的喜好、技术背景来选择。

后端我选择 Go, 因为用 Go 做的小工具（小网站）是最容易发布的, 现在做项目通常都会使用 GitHub 之类的代码仓库，而仓库地址就是 Go 项目的发布渠道，完全不需要额外的操作，而且又能打包为一个绿色的可执行文件。

再加上 Go 占用的资源极少，我只需要一个最低配置的 VPS 就能同时运行一大堆小工具。

容易发布/部署、运行资源少、编译速度快、web框架很轻但也够用，同时具备这几大优点的语言也只有 Go 了吧？（当然，如果做大项目, Go 的优势就不明显了。）

## 总设计（需求与灵感的细化、具体化）

对于我来说，总设计是最关键、最难，通常也是要花最多时间的一步。

总设计的本质是做选择。一个表面简单的需求背后通常隐藏在大量选择，比如：

- 做成单用户还是多用户系统？
- 只能传输文本吗，要不要提供传输图片或文件的功能？
- 每条消息的上限是多少字？
- 全部消息要长期保存吗，要做搜索功能吗？
- 要做 CLI 吗，网页版的功能与 CLI 有什么区别？
- 是完全公开还是要密码登入? 使用 CLI 时如何输入密码（每次输入还是缓存在哪里）？
- 每条消息可单独编辑吗，还是不允许编辑只能删除？

…… 等等。这只是一个极简单的小工具，要思考的东西算很少了，项目稍大一点要思考的因素就会指数式增长。

有人习惯先做，一边做一边想一边改，这个方法完全没有问题。我的习惯是先总体上尽量多想，让脑子里“成品”的形象越来越清晰，清晰到一定程度之后我才开始写代码。

## 起名，新建仓库

起名是个难点，我也没啥好办法，比较随意，比如这次这个项目我起名为 txt, 因为我决定只处理纯文本，不管图片和文件了。

起名后我习惯先去 GitHub 网页版新建仓库，然后用 GitHub Desktop 克隆到本地，再用 go mod init 命令新建一个 Go 项目。

接下来就可以开始写代码了。

## 数据模型/数据库

有人喜欢先写前端页面，先填充 mock 数据，然后再写后端增删改查。这当然是一种很合理的流程。

我喜欢先写数据模型或数据库的 create table 语句，而前端页面我可能先画草图，也可能只在脑中想象。

具体到我正在写的这个小工具，我会新建一个文件夹 model, 在里面新建文件 model.go, 内容为：

```go
package model

type TxtMsg struct {
	ID    string
	Msg   string
	CTime int64
	MTime int64
}
```

然后写与这个数据模型对应的数据表结构，新建一个文件夹 stmt, 在里面新建文件 stmt.go, 内容为：

```go
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
`
```

最后新建文件夹 mydb, 在里面新建文件 mydb.go, 内容为：

```go
package mydb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Path string
	DB   *sql.DB
}

func (db *DB) Open(dbPath string) (err error) {
	db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1")
	return
}
```

至此，我们就拥有了一个可供增删改查的数据库了。

model.go, stmt.go, mydb.go 这几个文件，相当于把数据套了几层:

- 最底层是 model, 它提供了最基本的积木，定义了积木的具体形状
- 最顶层是 mydb, 后续我们会为它增加 db.Add, db.Search, db.Delete, db.GetAll 等等对外公开的方法，让 web 框架可以非常轻松地使用数据库。
- stmt 是中间层，如果使用 ORM
