// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";
import { CreateCopyComp, MsgItem, TxtMsg } from "./txtmsg-item.js";

interface Alias {
  ID: string;
  MsgID: string;
}

const Alerts = util.CreateAlerts();
const TextForCopy = CreateCopyComp();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [util.LinkElem("/", { text: "Home" }), span(" .. Search")],
});

const AliasList = cc("ul");
const AliasShowBtn = cc("a", { attr: { href: "#" }, text: "(show)" });
const AliasHideBtn = cc("a", { attr: { href: "#" }, text: "(hide)" });
const Loading = util.CreateLoading();
const AliasesArea = cc("div", {
  children: [
    m("h4").text("Aliases").addClass("mb-0"),
    m("hr").addClass("my-0"),
    m("div")
      .addClass("text-right")
      .append(
        m(AliasShowBtn).on("click", toggleAlias).hide(),
        m(AliasHideBtn).on("click", toggleAlias)
      ),
    m(AliasList),
    m(Loading),
  ],
});

function toggleAlias(e: JQuery.ClickEvent): void {
  e?.preventDefault();
  AliasList.elem().toggle();
  AliasShowBtn.elem().toggle();
  AliasHideBtn.elem().toggle();
}
function hideAlias(): void {
  AliasList.hide();
  AliasShowBtn.show();
  AliasHideBtn.hide();
}

const MsgList = cc("div");

const SearchInput = util.create_input();
const SearchBtn = cc("button", { text: "Search" });
const SearchAlerts = util.CreateAlerts(2);

var firstSearch = true;

const Form = cc("form", {
  children: [
    util.create_item(SearchInput, "Search text", "", "mb-1"),
    m(SearchAlerts),
    m("div")
      .addClass("text-right")
      .append(
        m(SearchBtn).on("click", (e) => {
          e.preventDefault();
          const body = {
            keyword: util.val(SearchInput, "trim"),
            buckets: [],
          };
          if (!body.keyword) {
            util.focus(SearchInput);
            return;
          }
          SearchAlerts.insert("primary", `Searching [${body.keyword}]...`);
          util.ajax(
            {
              method: "POST",
              url: "/api/search",
              alerts: SearchAlerts,
              buttonID: SearchBtn.id,
              contentType: "json",
              body: body,
            },
            (resp) => {
              const items = resp as TxtMsg[];
              if (items && items.length > 0) {
                if (firstSearch) {
                  firstSearch = false;
                  hideAlias();
                }
                SearchAlerts.insert(
                  "success",
                  "Found " + items.length + " items."
                );
                MsgList.elem().empty();
                items.sort((a, b) => b.ID.localeCompare(a.ID));
                appendToList(MsgList, items.map(MsgItem));
              } else {
                SearchAlerts.insert("danger", "No items found.");
              }
            }
          );
        })
      ),
  ],
});

$("#root").append(
  m(NaviBar).addClass("my-5"),
  m(Form),
  m(Alerts),
  m(MsgList).addClass("mb-5"),
  m(AliasesArea).addClass("my-5"),
  m("div").text(".").addClass("Footer"),
  m(TextForCopy).hide()
);

init();

function init() {
  getAllAliases();
}

function getAllAliases(): void {
  util.ajax(
    { method: "GET", url: "/api/get-all-aliases", alerts: Alerts },
    (resp) => {
      const aliases = resp as Alias[];
      if (aliases && aliases.length > 0) {
        aliases.sort((a, b) => b.MsgID.localeCompare(a.MsgID));
        appendToList(AliasList, aliases.map(AliasItem));
      } else {
        AliasesArea.elem().hide();
      }
    },
    undefined,
    () => {
      Loading.hide();
      util.focus(SearchInput);
    }
  );
}

function AliasItem(alias: Alias): mjComponent {
  return cc("li", {
    children: [
      span(`[${alias.MsgID}]`).addClass("text-grey mr-2"),
      util.LinkElem("/public/edit.html?id=" + alias.MsgID, {
        text: alias.ID,
        blank: true,
      }),
    ],
  });
}
