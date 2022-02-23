// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";
import { CreateCopyComp, MsgItem } from "./txtmsg-item.js";
const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const TextForCopy = CreateCopyComp();
const NaviBar = cc("div", {
    classes: "my-5",
    children: [util.LinkElem("/", { text: "Home" }), span(" .. Search")],
});
const AliasList = cc("ul");
const AliasShowBtn = cc("a", { attr: { href: "#" }, text: "(show)" });
const AliasHideBtn = cc("a", { attr: { href: "#" }, text: "(hide)" });
const AliasesArea = cc("div", {
    children: [
        m("h4").text("Aliases").addClass("mb-0"),
        m("hr").addClass("my-0"),
        m("div")
            .addClass("text-right")
            .append(m(AliasShowBtn).on("click", toggleAlias).hide(), m(AliasHideBtn).on("click", toggleAlias)),
        m(AliasList),
    ],
});
function toggleAlias(e) {
    e === null || e === void 0 ? void 0 : e.preventDefault();
    AliasList.elem().toggle();
    AliasShowBtn.elem().toggle();
    AliasHideBtn.elem().toggle();
}
const MsgList = cc("div");
const SearchInput = util.create_input();
const SearchBtn = cc("button", { text: "Search" });
const FormAlerts = util.CreateAlerts();
var firstSearch = true;
const Form = cc("form", {
    children: [
        util.create_item(SearchInput, "Search text", "", "mb-1"),
        m(FormAlerts),
        m("div")
            .addClass("text-right")
            .append(m(SearchBtn).on("click", (e) => {
            e.preventDefault();
            if (firstSearch) {
                firstSearch = false;
                toggleAlias();
            }
            const body = {
                keyword: util.val(SearchInput, "trim"),
                buckets: [],
            };
            FormAlerts.insert("primary", `Searching [${body.keyword}]...`);
            util.ajax({
                method: "POST",
                url: "/api/search",
                alerts: FormAlerts,
                buttonID: SearchBtn.id,
                contentType: "json",
                body: body,
            }, (resp) => {
                const items = resp;
                if (items && items.length > 0) {
                    FormAlerts.insert("success", "Found " + items.length + " items.");
                    MsgList.elem().empty();
                    appendToList(MsgList, items.map(MsgItem));
                }
                else {
                    FormAlerts.insert("danger", "No items found.");
                }
            });
        })),
    ],
});
$("#root").append(m(NaviBar).addClass("my-5"), m(Loading).addClass("my-5"), m(Form).hide(), m(Alerts), m(MsgList).addClass("mb-5"), m(AliasesArea).addClass("my-5").hide(), m("div").text(".").addClass("Footer"), m(TextForCopy).hide());
init();
function init() {
    getAllAliases();
}
function AliasItem(alias) {
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
function getAllAliases() {
    util.ajax({ method: "GET", url: "/api/get-all-aliases", alerts: Alerts }, (resp) => {
        const aliases = resp;
        aliases.sort((a, b) => b.MsgID.localeCompare(a.MsgID));
        Form.show();
        AliasesArea.show();
        if (aliases && aliases.length > 0) {
            appendToList(AliasList, aliases.map(AliasItem));
        }
        else {
            AliasList.elem().append(m("li").text("No aliases found."));
        }
    }, undefined, () => {
        Loading.hide();
        util.focus(SearchInput);
    });
}
