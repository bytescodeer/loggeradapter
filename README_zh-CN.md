# loggeradapter

[English Doc](README.md) | [中文文档](README_zh-CN.md)

Golang 日志适配器，实现了日志轮转、备份和压缩归档、及按指定策略删除备份和归档文件的功能，可与常见日志组件如：zap logger、logrus 等结合。
注意：由于包路径更改的原因，版本号必须大于等于 v0.0.6！

## 使用示例

-   导入包

```go
  go get -u github.com/bytescodeer/loggeradapter
```

-   与 zap logger 结合

```go
  zapcore.AddSync(loggeradapter.New(loggeradapter.Config{
	Filename: "logs/log.log",
	Rotation: "50mb",
	Backup:   "1w",
	Archive:  "1M",
  }))
```

-   与 logrus 结合

```go
  logrus.SetOutput(loggeradapter.New(loggeradapter.Config{
    Filename: "logs/log.log",
    Rotation: "50mb",
    Backup:   "1w",
    Archive:  "1M",
  }))
```

## 参数说明

-   Filename

    指定日志输出文件路径及名称，不提供时默认为当前目录下 `./logs/log.log`。

    最新日志始终输出到该文件中，当该文件内容达到指定的最大文件大小或到达指定时间时（由参数 `Rotation` 指定的轮转策略确定），
    将被重命名为备份文件（备份文件名称由参数 `Backup` 确定），然后生成一个相同名称的文件，并将最新日志输出到该新文件中。

-   Rotation

    文件轮转备份策略，支持以下格式的配置：

    (1). `b|byte|kb|kilobyte|mb|megabyte|gb|gigabyte|tb|terabyte`

    (2). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (3). `annually|monthly|weekly|daily|hourly|minutely|secondly`

    _解释_： 以上三种配置方式中，`(1)` 和 `(2)` 配置必须前面有数字，如：10mb、2year，且配置值不区分大小写。其中：
    `b|byte|kb|kilobyte|mb|megabyte|gb|gigabyte|tb|terabyte`
    为按照文件大小轮转备份日志，如：10b、10byte、10Byte、10BYTE 都为文件大小达到 10 字节时轮转备份新日志文件，最大支持到 TB 级别。
    此时日志备份文件名称为指定的 `Filename` 连接上 `2006-01-02T15-04-05.000` 格式，如：`logs/log-2024-01-01T10-10-10.123.log`。
    `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`
    为按照时间轮转备份日志，如：1y、1year、1YEAR、1Year 都为 1 年生成一个新的备份文件，此时备份文件名称类似： `logs/log-2024.log`。
    比如指定为 1h、1hour、1HOUR、1Hour，则 1 小时生成一个新的备份文件，此时备份文件名称类似：`logs/log-2024-01-01T10.log`。
    `annually|monthly|weekly|daily|hourly|minutely|secondly`
    为按照时间周期轮转备份日志，如设置为 `annually`或 `Annually` 时，则每 1 年生成一个新的备份文件，此时和 1y、1year、1YEAR、1Year 具有相同的作用，其生成的文件名称类似：`logs/log-2024.log`。
    若设置为 `monthly` 或 `Monthly` 时，则每 1 月生成一个新的备份文件，此时和 1M、1month、1mo、1mon 具有相同的作用，其生成的文件名称类似：`logs/log-2024-01.log`。
    _注意_：M 为月，m 为分钟！`M|month|mo|mon` 都为月，`m|minute|min` 都为分钟！

-   Backup

    备份文件保留策略，支持以下格式的配置：

    (1). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (2). `number`

    _解释_： 以上两种配置方式中，`(1)` 配置和 `Rotation` 的 `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second` 配置方式相同。
    意思为按照周期保留备份文件，达到该时长的备份文件将被压缩并归档（若指定了压缩归档策略），同时删除旧备份文件。
    如指定为：1w、1week、1WEEK、1Week，则备份文件会保留 1 周，然后压缩并归档，同时删除旧备份文件。
    以上配置 `(2)` 为指定单个数字，如设置为：10，则表示备份文件达到 10 个时，将压缩并归档，同时删除旧备份文件。

-   Archive

    压缩归档策略，支持以下格式的配置：

    (1). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (2). `number`

    _解释_： 以上两种配置方式和 `Backup` 相同。为 `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second` 时按照周期保留压缩归档文件，达到该时长的压缩归档文件将被删除。
    如指定为：`1M、1month、mo、mon`，则压缩归档文件会保留 1 个月，然后删除 1 个月之前的所有压缩归档文件。
    若指定为数字 `number` 时，则会最多保留该数量的压缩归档文件，多余的压缩归档文件将被删除。如指定为 10，则最多保留 10 个最新的压缩归档文件，其余的压缩归档文件将被删除。

    _注意_：当该参数为空时，则不压缩归档文件，只会按照保留策略删除符合条件的旧文件！

## 注意事项

若参数 `Rotation`、`Backup`、`Archive` 三个参数都为空时，则日志文件将不会轮转备份，也不会进行压缩归档，日志会持续不断的输出到指定日志文件中。

## 编写初衷

在实际项目开发中，使用过 zap logger、logrus、glog、zerolog 等日志组件，发现日志组件的日志轮转备份功能不足， 一般都是通过集成 `gopkg.in/natefinch/lumberjack.v2` 做日志轮转备份，而该组件只支持按照大小轮转备份，实际开发当中有按照时间周期轮转备份的需求，就需要自己编写实现，且配置不够灵活，无法满足实际需求，故自己做了封装实现，以便在其他不同项目中能够直接使用。

## 其他

欢迎各位大佬提意见、Issues、PR！🤝👊🫶
