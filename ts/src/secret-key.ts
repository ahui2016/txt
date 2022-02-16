// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span } from "./mj.js";
import * as util from "./util.js";

const FormAlerts = util.CreateAlerts();
const footerElem = util.CreateFooter();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [
    util.LinkElem("/", { text: "home" }),
    span(" .. "),
    util.LinkElem("/public/sign-in.html", { text: "Sign-in" }),
    span(" .. Secret Key"),
  ],
});

const aboutPage = m("div").append(
  m("p").text(
    "输入主密码，点击 Get Key 按钮可获取当前密钥。" +
      "获取密钥后会出现 Generate 按钮，点击该按钮可生成新的密钥。" +
      "一旦生成新密钥，旧密钥就会作废。"
  )
);

const CurrentKeyArea = cc("div");
CurrentKeyArea.init = function (key: util.CurrentKey) {
  const self = CurrentKeyArea.elem();
  self.append(
    m("div").append(
      span("Current Key"),
      m("input").addClass("ml-2").val(key.Key)
    )
  );
  self.append(
    m("div").text(`生效日期: ${dayjs.unix(key.Starts).format("YYYY-MM-DD")}`)
  );
  self.append(m("div").text(`有效期: ${key.MaxAge} (天)`));
  if (key.IsGood) {
    self.append(
      m("div").append(span("状态: "), span("有效").addClass("alert-success"))
    );
    self.append(
      m("div")
        .addClass("form-text")
        .text(
          `(该密钥将于 ${dayjs.unix(key.Expires).format("YYYY-MM-DD")} 自动作废)`
        )
    );
  } else {
    self.append(
      m("div").append(span("状态: "), span("已失效").addClass("alert-danger"))
    );
    self.append(
      m("div").text(
        `该密钥已于 ${dayjs.unix(key.Expires).format("YYYYMMDD")} 作废`
      )
    );
  }
};

// https://www.chromium.org/developers/design-documents/form-styles-that-chromium-understands/
const UsernameInput = cc("input", { attr: { autocomplete: "username" } });
const PwdInput = cc("input", { attr: { autocomplete: "current-password" } });
const GetKeyBtn = cc("button", { text: "Get key" });
const GenKeyBtn = cc("button", { text: "Generate" });

const Form = cc("form", {
  children: [
    m("label").text("Master Password").attr({ for: PwdInput.raw_id }),
    m("div").append(
      m(UsernameInput).hide(),
      m(PwdInput)
        .addClass("form-textinput form-textinput-fat")
        .attr({ type: "password" }),
      m("div").append(
        m(GetKeyBtn).on("click", (event) => {
          event.preventDefault();
          const pwd = util.val(PwdInput);
          if (!pwd) {
            util.focus(PwdInput);
            return;
          }
          util.disable(GetKeyBtn);
          util.ajax(
            {
              method: "POST",
              url: "/auth/get-current-key",
              alerts: FormAlerts,
              body: { password: pwd },
            },
            // success
            (resp) => {
              const currentKey = resp as util.CurrentKey;
              CurrentKeyArea.elem().show();
              CurrentKeyArea.init!(currentKey);
              GenKeyBtn.elem().show();
              FormAlerts.clear();
            },
            // fail
            (_, errMsg) => {
              util.enable(GetKeyBtn);
              FormAlerts.insert("danger", errMsg);
              util.focus(PwdInput);
            }
          );
        }),
        m(GenKeyBtn)
          .addClass("ml-2")
          .on("click", (event) => {
            event.preventDefault();
            const pwd = util.val(PwdInput);
            if (!pwd) {
              util.focus(PwdInput);
              return;
            }
            util.ajax(
              {
                method: "POST",
                url: "/auth/gen-new-key",
                alerts: FormAlerts,
                buttonID: GenKeyBtn.id,
                body: { password: pwd },
              },
              // success
              (resp) => {
                const currentKey = resp as util.CurrentKey;
                CurrentKeyArea.elem().html("");
                CurrentKeyArea.init!(currentKey);
                PwdInput.elem().val("");
              }
            );
          })
          .hide()
      )
    ),
  ],
});

$("#root").append(
  m(NaviBar),
  aboutPage,
  m(Form),
  m(FormAlerts),
  m(CurrentKeyArea).addClass("my-5").hide(),
  footerElem.hide()
);

init();

function init() {
  $("title").text("Secret Key .. txt");
  util.focus(PwdInput);
}
