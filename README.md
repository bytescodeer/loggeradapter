# loggeradapter

[English Doc](README.md) | [中文文档](README_zh-CN.md)

The Go log adapter implements log rotation, backup, compression archiving,
and the functionality to delete backup and archived files according to specified policies.
It can be integrated with common logging components such as zap logger, logrus, etc.
Note: Due to the change of the package path, the version number must be greater than or equal to v0.0.6!

## Usage

-   Import package

```go
  go get -u github.com/bytescodeer/loggeradapter
```

-   Integrate with zap logger

```go
  zapcore.AddSync(loggeradapter.New(loggeradapter.Config{
	Filename: "logs/log.log",
	Rotation: "50mb",
	Backup:   "1w",
	Archive:  "1M",
  }))
```

-   Integrate with logrus

```go
  logrus.SetOutput(loggeradapter.New(loggeradapter.Config{
    Filename: "logs/log.log",
    Rotation: "50mb",
    Backup:   "1w",
    Archive:  "1M",
  }))
```

## Configuration Instructions

-   Filename

    Specify the log output file path and name. If not provided, the default is ./logs/log.log in the current directory.

    The latest logs will always be written to this file. When the file size reaches the specified maximum size or
    the specified time is reached (determined by the Rotation parameter), the file will be renamed as a backup file
    (the backup file name is determined by the Backup parameter). A new file with the same name will be created,
    and the latest logs will be written to this new file.

-   Rotation

    The file rotation backup strategy supports configurations in the following formats:

    (1). `b|byte|kb|kilobyte|mb|megabyte|gb|gigabyte|tb|terabyte`

    (2). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (3). `annually|monthly|weekly|daily|hourly|minutely|secondly`

    _Explanation_: Among the three configuration methods above, configurations `(1)` and `(2)` must include a number in front,
    such as: `10mb, 2year`, and the values are case-insensitive. Among them:
    `b|byte|kb|kilobyte|mb|megabyte|gb|gigabyte|tb|terabyte` are used for log rotation and backup based on file size.
    For example, `10b, 10byte, 10Byte, 10BYTE` all indicate that the log file will rotate and back up when the file size reaches 10 bytes.
    The maximum size supported is up to TB. In this case,
    the log backup file name will be the specified Filename followed by the timestamp in `2006-01-02T15-04-05.000` format,
    such as `logs/log-2024-01-01T10-10-10.123.log`.
    `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second` are used for log rotation and backup based on time. For example, `1y, 1year, 1YEAR, 1Year`
    all indicate that a new backup file will be generated after 1 year. The backup file name will be similar to `logs/log-2024.log`. For example,
    if set as `1h, 1hour, 1HOUR, 1Hour`, a new backup file will be generated every 1 hour,
    and the backup file name will be similar to `logs/log-2024-01-01T10.log`.
    `annually|monthly|weekly|daily|hourly|minutely|secondly` are used for log rotation and backup based on time intervals. For example,
    setting annually or Annually will generate a new backup file every 1 year, which has the same effect as `1y, 1year, 1YEAR, 1Year`,
    and the file name generated will be similar to `logs/log-2024.log`. If set as monthly or Monthly, a new backup file will be generated every 1 month,
    which has the same effect as `1M, 1month, 1mo, 1mon`, and the file name generated will be similar to `logs/log-2024-01.log`.

    _Notice_: `M` is for month, and `m` is for minute! `M|month|mo|mon` all represent month, and `m|minute|min all` represent minute!

-   Backup

    Backup file retention policy supports configuration in the following formats:

    (1). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (2). `number`

    _Explanation_: Among the above two configuration methods, `(1)` configuration and `Rotation` of
    `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s |second` is configured in the same way.
    This means that backup files are retained according to the period.
    Backup files that reach this length will be compressed and archived
    (if the compression archiving policy is specified), and old backup files will be deleted at the same time.
    If specified as: `1w, 1week, 1WEEK, 1Week`, the backup file will be retained for 1 week,
    then compressed and archived, and old backup files will be deleted.
    The above configuration `(2)` specifies a single number. If set to: 10,
    it means that when the number of backup files reaches 10,
    it will be compressed and archived, and old backup files will be deleted.

-   Archive

    Compression archiving strategy supports configuration in the following formats:

    (1). `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`

    (2). `number`

    _Explanation_： The above two configuration methods are the same as `Backup`.
    When it is `y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second`,
    compressed archive files are retained according to the period,
    and the compressed archives that reach this length are The file will be deleted.
    If specified as: `1M, 1month, mo, mon`, the compressed archive files will be retained for 1 month,
    and then all compressed archive files older than 1 month will be deleted.
    If specified as a number `number`, a maximum of this number of compressed archive files will be retained,
    and excess compressed archive files will be deleted. If specified as 10,
    a maximum of 10 latest compressed archive files will be retained and the remaining compressed archive files will be deleted.

    _Notice_：When this parameter is empty, the archive file will not be compressed, and only old files that meet the conditions will be deleted according to the retention policy!

## Things to note

If the three parameters `Rotation`, `Backup`, and `Archive` are all empty, the log file will not be rotated for backup,
nor will it be compressed and archived, and the log will be continuously output to the specified log file.

## Original intention of writing

In actual project development, I have used log components such as zap logger, logrus, glog, zerolog, etc.,
and found that the log rotation backup function of the log component is insufficient. Generally,
it is integrated through `gopkg.in/natefinch/lumberjack.v2` Do log rotation backup, and this component only supports rotation backup according to size.
In actual development, if there is a demand for rotation backup according to time period, you need to write the implementation yourself,
and the configuration is not flexible enough to meet the actual needs, so I made the encapsulation implementation myself.
So that it can be used directly in other different projects.

## Others

Welcome everyone to give your opinions、Issues、PR！🤝👊🫶
