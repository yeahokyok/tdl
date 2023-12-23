---
title: "迁移"
weight: 50
---

# 迁移

备份或恢复您的数据

## 备份

将您的数据备份到文件中。默认值：`<date>.backup.tdl`。

{{< command >}}
tdl backup
{{< /command >}}

或者指定输出文件：

{{< command >}}
tdl backup -d /path/to/custom.tdl
{{< /command >}}

## 恢复

从备份文件中恢复您的数据。

{{< command >}}
tdl recover -f /path/to/custom.backup.tdl
{{< /command >}}
