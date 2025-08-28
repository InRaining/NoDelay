# NoDelay

## 📝 本项目简介

- 该项目是一个基于ZBProxy魔改的Minecraft服务器代理程序。NoDelay是一个功能强大的代理工具，用于优化和管理Minecraft游戏的网络连接。

## 🧭 新增功能

⚙️ **更好的自定义配置**

- 该项目添加的外部白名单/黑名单获取支持，您现在可以通过API传入您需要添加的玩家。
- 该项目添加了私有配置，包括白名单API地址、自定义表头、联系名称和链接等。

```json
{
    "PrivateConfig": {
        "ListAPI": "http://bind.jsip.fun/isWhitelist.php",
        "Header": "HLN-Boost",
        "ContactName": "官方QQ售后群",
        "ContactLink": "666259678"
    }
}
```

- 白名单/黑名单配置说明：请确保您的玩家API能通过Get形式传入playerName参数，例如：`https://example.com/isWhitelist.php?playerName=`，ListAPI中不要带有`?playerName=`,并且当playerName正确或查询到的情况下，返回playerName。且在ListAPI设置完毕后，选择合适的模式，并对NoDelay进行冷重载，完成配置。

🔑 **启动验证**

- 该项目在代理的启动部分做了处理，将会在启动前检测是否具有白名单，以提高安全性。
- 验证配置说明：您可以任何形式设计您的验证方式，那取决于您的验证API，只需保证本程序访问您的API，最后返回的字符串为`true`或`false`即可。

🔨 **部分汉化支持**

- 该项目对部分返回信息如`玩家踢出显示`等进行了汉化(主要是`kick.go`中的内容)，但尚未对内端返回信息进行处理汉化，目前还在计划中。

⚡ **更多显示模式**

- 该项目新增了更多的模式，如下线模式`DownMode`，娱乐模式`JokeMode`等，以更好适配不同情形。
- 该项目新增了玩家第一次加入显示，以告知新玩家必要的信息。

## ❗️ 注意事项

- 请根据实际需求修改配置文件中的各项设置。
- 详细的配置说明和使用方法，请参考项目文档或相关资源。
- 本项目严格遵守`CC BY-NC-SA 4.0`协议开源，您可以自由地对本项目代码进行使用、编译、修改、分发等行为，但不享有其著作权，并不可将本项目用于任何直接或变相形式的商业用途，且任何形式使用、分发、修改等必须指明源项目或代码出处。如需用于商业用途，请联系本项目作者。

## ✨ 感谢使用

- 感谢您选择使用NoDelay代理程序，希望它能帮助您优化和管理Minecraft游戏的网络连接。如果您有任何建议或意见，欢迎随时反馈给我们。祝您游戏愉快！

## 🌐 参考链接

- ZBProxy 原作者: [Layou233](https://github.com/Layou233)
- NoDelay 作者: [InRaining](https://github.com/InRaining)
- NoDelay GitHub 项目页：[https://github.com/InRaining/NoDelay](https://github.com/InRaining/NoDelay)
