# SessionRepository：Robolectric vs 仪器测试 & 覆盖率对比（作业用）

## 两套测试分别是什么

| 测试类 | 路径 | 运行环境 | 作业里可称为 |
|--------|------|----------|----------------|
| `SessionRepositoryRobolectricTest` | `src/test` | 电脑 JVM + Robolectric 影子框架 | **Robolectric / JVM 单元测** |
| `SessionRepositoryInstrumentedTest` | `src/androidTest` | 真机或模拟器上的 **原生 Android** | **仪器测 / 原生环境** |

被测代码都是 `SessionRepository.kt`。两套用例 **场景镜像**（清会话、切换用户、循环写入、Flow 收集多条），便于对比「同逻辑」在不同运行栈下的覆盖率与稳定性。

## 如何跑

**Robolectric（无需设备）：**

```bash
./gradlew :app:testDebugUnitTest
```

结合 JaCoCo（工程里已有 `jacocoTestReport`）：

```bash
./gradlew :app:testDebugUnitTest :app:jacocoTestReport
# HTML: app/build/reports/jacoco/jacocoTestReport/html/index.html
```

**仪器测试（需已连接设备或已启动模拟器）：**

```bash
./gradlew :app:connectedDebugAndroidTest
```

开启 `debug.enableAndroidTestCoverage` 后，可生成仪器覆盖率（任务名随 AGP 可能为 `createDebugAndroidTestCoverageReport`，请以 `./gradlew :app:tasks --group reporting` 为准）。

## 如何对比覆盖率（作业表格示例）

1. 记 **基线**：只跑 Robolectric 相关 / 全量 unit test 的 JaCoCo 总览里，`SessionRepository` 行覆盖。
2. 再跑 **connected** 报告里同一文件的覆盖。
3. 对比差异时说明：两者都会触发 DataStore 读写，但 **环境不同**（真机/模拟器上的原生栈 vs JVM 影子框架），报告里对 **Framework / 依赖库** 的统计差异会很大；对比作业可关注 **`SessionRepository.kt` 自身行/分支覆盖**。

## 关于 `ResourcesImpl` / `ComplexColor` / `Bad return type`

这通常是 **Robolectric 与「更高 SDK 的资源实现 + 字节码插桩（含 IDE/Android Studio 的 With Coverage）」** 叠加导致的 **VerifyError**，不一定是「Robolectric 特别老」这一条原因，但 **升级到 Robolectric 4.16.x** 并 **固定影子 SDK** 往往更稳。

本工程已做：

- `gradle/libs.versions.toml`：`robolectric` **4.16**
- `app/src/test/resources/robolectric.properties`：`sdk=34`（与 `@Config(sdk = [34])` 一致，减少去拉 API 35 框架时踩坑）
- `debug.enableUnitTestCoverage` + `testOptions.unitTests.isIncludeAndroidResources`

若 **仅在 Android Studio 点 “Run with Coverage”** 仍崩溃，可改为：

1. 普通 Run（不带 IDE Coverage）执行 Robolectric；或  
2. 只用 Gradle：`./gradlew :app:testDebugUnitTest` / `jacocoTestReport`。

（社区里也有「JaCoCo/Kover 与 Robolectric 插桩冲突」的讨论；Gradle 报告与 IDE 内覆盖率不必完全一致。）

## 依赖版本

- Robolectric：见 `gradle/libs.versions.toml` 的 `robolectric`（当前为 **4.16**；若需可在 `4.16` / `4.16.1` 间微调）。
