// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { m, cc, span } from "./mj.js";
import * as util from "./util.js";
const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();
const NaviBar = cc("div", {
    classes: "my-5",
    children: [
        util.LinkElem("/", { text: "home" }),
        span(" .. "),
        util.LinkElem("/public/secret-key.html", { text: "get secret key" }),
        span(" .. Sign-in"),
    ],
});
const SignOutBtn = cc("button");
const SignOutArea = cc("div", {
    children: [
        m(SignOutBtn)
            .text("Sign out")
            .on("click", (event) => {
            event.preventDefault();
            util.ajax({
                method: "GET",
                url: "/auth/sign-out",
                alerts: Alerts,
                buttonID: SignOutBtn.id,
            }, () => {
                Alerts.clear().insert("info", "已登出");
                SignOutArea.elem().hide();
                SignInForm.elem().show();
                util.focus(PwdInput);
            });
        }),
    ],
});
const GotoGetKey = cc("div", { children: [
        span("获取密钥或重新生成密钥 ➡ "),
        util.LinkElem("/public/secret-key.html")
    ] });
// https://www.chromium.org/developers/design-documents/form-styles-that-chromium-understands/
const UsernameInput = cc("input", { attr: { autocomplete: "username" } });
const PwdInput = cc("input", { attr: { autocomplete: "current-password" } });
const SubmitBtn = cc("button", { text: "Sign in" });
const SignInForm = cc("form", {
    children: [
        m("label").text("Secret Key").attr({ for: PwdInput.raw_id }),
        m("div").append(m(UsernameInput).hide(), m(PwdInput).attr({ type: "password" }), m(SubmitBtn)
            .addClass("ml-1")
            .on("click", (event) => {
            event.preventDefault();
            const pwd = util.val(PwdInput);
            if (!pwd) {
                util.focus(PwdInput);
                return;
            }
            util.ajax({
                method: "POST",
                url: "/auth/sign-in",
                alerts: Alerts,
                buttonID: SubmitBtn.id,
                body: { password: pwd },
            }, () => {
                PwdInput.elem().val("");
                SignInForm.elem().hide();
                Alerts.clear().insert("success", "成功登入");
                SignOutArea.elem().show();
                GotoGetKey.elem().hide();
            }, (that, errMsg) => {
                if (that.status == 401) {
                    Alerts.insert("danger", "密码错误");
                    GotoGetKey.elem().show();
                }
                else {
                    Alerts.insert("danger", errMsg);
                }
            }, () => {
                util.focus(PwdInput);
            });
        })),
    ],
});
$("#root").append(m(NaviBar), m(Loading).addClass("my-3"), m(SignInForm).hide(), m(Alerts), m(GotoGetKey).hide(), m(SignOutArea).addClass("my-5").hide(), footerElem.hide());
init();
function init() {
    $('title').text("Sign-in .. txt");
    checkSignIn();
}
function checkSignIn() {
    util.ajax({ method: "GET", url: "/auth/is-signed-in", alerts: Alerts }, (resp) => {
        const yes = resp;
        if (yes) {
            Alerts.insert("info", "已登入");
            SignOutArea.elem().show();
            Loading.hide();
        }
        else {
            SignInForm.elem().show();
            util.focus(PwdInput);
        }
    }, undefined, () => {
        Loading.hide();
    });
}
