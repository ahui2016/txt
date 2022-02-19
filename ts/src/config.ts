// 采用受 Mithril 启发的基于 jQuery 实现的极简框架 https://github.com/ahui2016/mj.js
import { mjElement, mjComponent, m, cc, span, appendToList } from "./mj.js";
import * as util from "./util.js";

const Alerts = util.CreateAlerts();
const Loading = util.CreateLoading("center");
const footerElem = util.CreateFooter();

const GotoSignIn = util.CreateGotoSignIn();

const NaviBar = cc("div", {
  classes: "my-5",
  children: [
    util.LinkElem("/", { text: "home" }),
    span(" .. "),
    span(" .. Config"),
  ],
});

// type Config struct {
// 	Password       string // 主密码，唯一作用是生成 Key
// 	Key            string // 日常使用的密钥
// 	KeyStarts      int64  // Key 的生效时间 (timestamp)
// 	KeyMaxAge      int64  // Key 的有效期（秒）
// 	MsgSizeLimit   int64  // 每条消息的长度上限
// 	TempLimit      int64  // 暂存消息条数上限（永久消息不设上限）
// 	EveryPageLimit int64  // 每页最多列出多少条消息
// 	TimeOffset     string // "+8" 表示北京时间, "-5" 表示纽约时间, 依此类推。
// }

const MaxAgeInput = util.create_input();
const MsgSizeInput = util.create_input();
const TempLimitInput = util.create_input();
const PageLimitInput = util.create_input();
const TimezoneInput = util.create_input();
const FormAlerts = util.CreateAlerts();
const HiddenBtn = cc("button", { id: "submit", text: "submit" }); // 这个按钮是隐藏不用的，为了防止按回车键提交表单
const SubmitBtn = cc("button", { text: "Submit" });

const Form = cc("form", {
  children: [
    util.create_item(MaxAgeInput, "Key Max Age", "密钥有效期（单位：天）"),
    util.create_item(
      MsgSizeInput,
      "Message Size Limit",
      "每条消息的长度上限 (单位：byte)"
    ),
    util.create_item(
      TempLimitInput,
      "Temporary Messages Limit",
      "暂存消息条数上限"
    ),
    util.create_item(PageLimitInput, "Page Limit", "每页最多列出多少条消息"),
    util.create_item(
      TimezoneInput,
      "Timezone",
      '时区（例如 "+8" 表示北京时间, "-5" 表示纽约时间）'
    ),
    m(FormAlerts),
    m(HiddenBtn)
      .hide()
      .on("click", (e) => {
        e.preventDefault();
        return false; // 这个按钮是隐藏不用的，为了防止按回车键提交表单。
      }),
    m(SubmitBtn).on("click", (e) => {
      e.preventDefault();
      const body = {
        KeyMaxAge: util.val(MaxAgeInput, "trim"),
        MsgSizeLimit: util.val(MsgSizeInput, "trim"),
        TempLimit: util.val(TempLimitInput, "trim"),
        EveryPageLimit: util.val(PageLimitInput, "trim"),
        TimeOffset: util.val(TimezoneInput, "trim"),
      };
      util.ajax({
        method: "POST",
        url: "/api/update-config",
        alerts: FormAlerts,
        buttonID: SubmitBtn.id,
        contentType: "json",
        body: body,
      });
    }),
  ],
});
