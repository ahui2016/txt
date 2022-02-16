// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { m, cc } from "./mj.js";
import * as util from "./util.js";
// import { MimaItem } from "./mima-item.js";
const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();
const titleArea = m("div").addClass("text-center").append(m("h1").text("txt"));
const GotoSignIn = util.CreateGotoSignIn();
const MimaList = cc("div");
const TextForCopy = cc("input", { id: "TextForCopy" });
const MsgInput = util.create_textarea();
const SendBtn = cc("button", { text: "Send" });
const FormAlerts = util.CreateAlerts();
const Form = cc("form", {
    children: [
        m(MsgInput).addClass("form-textinput form-textinput-fat"),
        m("div")
            .addClass("text-right")
            .append(m(SendBtn).on("click", (e) => {
            e.preventDefault();
            util.ajax({
                method: "POST",
                url: "/api/add",
                alerts: FormAlerts,
                buttonID: SendBtn.id,
                body: { msg: util.val(MsgInput, "trim") },
            }, (resp) => {
                const id = resp.message;
                FormAlerts.insert("success", id);
                MsgInput.elem().val("");
                util.focus(MsgInput);
            });
        })),
        m(FormAlerts),
    ],
});
$("#root").append(titleArea, m(Loading).addClass("my-3"), m(Alerts), m(GotoSignIn).addClass("my-3").hide(), m(Form).hide(), m(MimaList).addClass("mt-3"), footerElem.hide(), m(TextForCopy).hide());
init();
function init() {
    checkSignIn();
}
function checkSignIn() {
    util.ajax({ method: "GET", url: "/auth/is-signed-in", alerts: Alerts }, (resp) => {
        const yes = resp;
        if (yes) {
            Form.elem().show();
            util.focus(MsgInput);
        }
        else {
            GotoSignIn.elem().show();
        }
    }, undefined, () => {
        Loading.hide();
    });
}
