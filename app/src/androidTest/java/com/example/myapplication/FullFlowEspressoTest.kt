package com.example.myapplication

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.hasClickAction
import androidx.compose.ui.test.hasSetTextAction
import androidx.compose.ui.test.hasText
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performTextInput
import androidx.test.espresso.Espresso
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

/**
 * 全流程 UI 集成测试。
 * Compose Test API + Espresso 空闲同步，真机/模拟器运行。
 * 数据通过 Room + MockRemoteApi，不需要后端。
 */
@RunWith(AndroidJUnit4::class)
class FullFlowEspressoTest {

    @get:Rule
    val composeRule = createAndroidComposeRule<MainActivity>()

    private fun app() = InstrumentationRegistry.getInstrumentation()
        .targetContext.applicationContext as CampusApplication

    private val ts get() = System.currentTimeMillis()

    @Before
    fun setUp() = runBlocking {
        // 清除登录状态 + 清空 Room 数据库，避免残留数据干扰
        app().container.auth.logout()
        app().container.db.clearAllTables()
    }

    @After
    fun tearDown() = runBlocking { app().container.auth.logout() }

    // ════════════════════════════════════════
    //  辅助
    // ════════════════════════════════════════

    /** 精确匹配可点击的按钮/文字（避免"登录"标题 vs 按钮歧义） */
    private fun clickBtn(text: String) {
        composeRule.onNode(hasText(text) and hasClickAction()).performClick()
        Espresso.onIdle()
    }

    /** 点击底部 Tab（NavigationBar 里的 label） */
    private fun clickTab(label: String) = clickBtn(label)

    /** 注册新用户并进入主界面 */
    private fun registerAndEnterMain(
        sid: String = "e2e_$ts",
        pwd: String = "test1234",
        nick: String = "测试用户",
    ) {
        clickBtn("没有账号？去注册")

        composeRule.onNode(hasSetTextAction() and hasText("学号")).performTextInput(sid)
        composeRule.onNode(hasSetTextAction() and hasText("密码")).performTextInput(pwd)
        composeRule.onNode(hasSetTextAction() and hasText("昵称（可选）")).performTextInput(nick)

        clickBtn("注册并登录")

        composeRule.waitUntil(10_000) {
            composeRule.onAllNodes(hasText("校园二手")).fetchSemanticsNodes().isNotEmpty()
        }
    }

    /** 发布一个商品（在发布页） */
    private fun publishProduct(title: String, desc: String, price: String, category: String = "教材图书") {
        composeRule.onNode(hasSetTextAction() and hasText("标题")).performTextInput(title)
        composeRule.onNode(hasSetTextAction() and hasText("描述")).performTextInput(desc)
        composeRule.onNode(hasSetTextAction() and hasText("价格（元）")).performTextInput(price)
        clickBtn("提交")
        composeRule.waitUntil(5_000) {
            composeRule.onAllNodes(hasText("发布成功")).fetchSemanticsNodes().isNotEmpty()
        }
    }

    /** 等待某个文本出现 */
    private fun waitForText(text: String, timeout: Long = 5_000) {
        composeRule.waitUntil(timeout) {
            composeRule.onAllNodes(hasText(text)).fetchSemanticsNodes().isNotEmpty()
        }
    }

    // ════════════════════════════════════════
    //  1. 认证
    // ════════════════════════════════════════

    @Test
    fun auth_loginScreen_showsFields() {
        // "登录"同时出现在标题和按钮上，用 hasClickAction 区分按钮
        composeRule.onNode(hasText("登录") and hasClickAction()).assertIsDisplayed()
        composeRule.onNode(hasSetTextAction() and hasText("学号")).assertIsDisplayed()
        composeRule.onNode(hasSetTextAction() and hasText("密码")).assertIsDisplayed()
        composeRule.onNodeWithText("没有账号？去注册").assertIsDisplayed()
    }

    @Test
    fun auth_navigateToRegister_andBack() {
        clickBtn("没有账号？去注册")
        composeRule.onNodeWithText("注册").assertIsDisplayed()
        composeRule.onNodeWithText("昵称（可选）").assertIsDisplayed()

        clickBtn("返回登录")
        composeRule.onNodeWithText("没有账号？去注册").assertIsDisplayed()
    }

    @Test
    fun auth_registerAndLogin_showsMainShell() {
        registerAndEnterMain()
        // 底部 Tab 用 hasClickAction 精确匹配
        composeRule.onNode(hasText("首页") and hasClickAction()).assertIsDisplayed()
        composeRule.onNode(hasText("发布") and hasClickAction()).assertIsDisplayed()
        composeRule.onNode(hasText("收藏") and hasClickAction()).assertIsDisplayed()
        composeRule.onNode(hasText("消息") and hasClickAction()).assertIsDisplayed()
        composeRule.onNode(hasText("我的") and hasClickAction()).assertIsDisplayed()
    }

    @Test
    fun auth_wrongPassword_showsError() {
        val sid = "err_$ts"
        registerAndEnterMain(sid = sid)

        clickTab("我的")
        clickBtn("退出登录")
        waitForText("没有账号？去注册")

        // 错误密码登录
        composeRule.onNode(hasSetTextAction() and hasText("学号")).performTextInput(sid)
        composeRule.onNode(hasSetTextAction() and hasText("密码")).performTextInput("wrong")
        clickBtn("登录")
        waitForText("学号或密码错误")
        composeRule.onNodeWithText("学号或密码错误").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  2. 首页
    // ════════════════════════════════════════

    @Test
    fun home_showsSearchBarAndVisibleCategories() {
        registerAndEnterMain()
        composeRule.onNodeWithText("搜索标题或描述").assertIsDisplayed()
        // 只检查前几个可见分类（"运动"等可能在横向滚动外）
        composeRule.onNodeWithText("全部分类").assertIsDisplayed()
        composeRule.onNodeWithText("教材图书").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  3. 发布
    // ════════════════════════════════════════

    @Test
    fun publish_fillForm_success() {
        registerAndEnterMain()
        clickTab("发布")
        publishProduct("测试书本", "九成新教材", "25.5")
        composeRule.onNodeWithText("发布成功").assertIsDisplayed()
    }

    @Test
    fun publish_invalidPrice_showsError() {
        registerAndEnterMain()
        clickTab("发布")
        composeRule.onNode(hasSetTextAction() and hasText("标题")).performTextInput("坏价格")
        composeRule.onNode(hasSetTextAction() and hasText("描述")).performTextInput("desc")
        composeRule.onNode(hasSetTextAction() and hasText("价格（元）")).performTextInput("abc")
        clickBtn("提交")
        waitForText("价格格式不正确")
        composeRule.onNodeWithText("价格格式不正确").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  4. 商品详情 + 收藏
    // ════════════════════════════════════════

    @Test
    fun product_publishAndViewDetail() {
        registerAndEnterMain()
        clickTab("发布")
        publishProduct("详情测试商品", "这是描述", "10")

        clickTab("我的")
        clickBtn("我的发布")
        waitForText("详情测试商品")
        clickBtn("详情测试商品")

        waitForText("商品详情")
        composeRule.onNodeWithText("商品详情").assertIsDisplayed()
        composeRule.onNodeWithText("在售").assertIsDisplayed()
    }

    @Test
    fun favorites_empty_showsPlaceholder() {
        registerAndEnterMain()
        clickTab("收藏")
        waitForText("还没有收藏商品")
        composeRule.onNodeWithText("还没有收藏商品").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  5. 消息
    // ════════════════════════════════════════

    @Test
    fun chat_empty_showsPlaceholder() {
        registerAndEnterMain()
        clickTab("消息")
        waitForText("暂无消息")
        composeRule.onNodeWithText("暂无消息").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  6. 个人中心
    // ════════════════════════════════════════

    @Test
    fun profile_showsUserInfoAndActions() {
        registerAndEnterMain(nick = "E2E小明")
        clickTab("我的")
        waitForText("E2E小明")
        composeRule.onNodeWithText("E2E小明").assertIsDisplayed()
        composeRule.onNodeWithText("编辑资料").assertIsDisplayed()
        composeRule.onNodeWithText("我的订单").assertIsDisplayed()
        composeRule.onNodeWithText("我的发布").assertIsDisplayed()
        composeRule.onNodeWithText("退出登录").assertIsDisplayed()
    }

    @Test
    fun profile_editProfile_showsForm() {
        registerAndEnterMain(nick = "原始昵称")
        clickTab("我的")
        waitForText("编辑资料")
        clickBtn("编辑资料")

        // 等表单加载完成后检查
        waitForText("昵称")
        composeRule.onNode(hasSetTextAction() and hasText("昵称")).assertIsDisplayed()
        composeRule.onNode(hasSetTextAction() and hasText("电话（可选）")).assertIsDisplayed()
        composeRule.onNode(hasText("保存") and hasClickAction()).assertIsDisplayed()
    }

    @Test
    fun profile_myOrders_empty() {
        registerAndEnterMain()
        clickTab("我的")
        waitForText("我的订单")
        clickBtn("我的订单")
        waitForText("暂无订单")
        composeRule.onNodeWithText("暂无订单").assertIsDisplayed()
    }

    @Test
    fun profile_myPublished_empty() {
        registerAndEnterMain()
        clickTab("我的")
        waitForText("我的发布")
        clickBtn("我的发布")
        waitForText("还没有发布商品")
        composeRule.onNodeWithText("还没有发布商品").assertIsDisplayed()
    }

    @Test
    fun profile_notifications_empty() {
        registerAndEnterMain()
        clickTab("我的")
        waitForText("通知")
        clickBtn("通知")
        waitForText("暂无通知")
        composeRule.onNodeWithText("暂无通知").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  7. 退出登录
    // ════════════════════════════════════════

    @Test
    fun logout_returnsToLoginScreen() {
        registerAndEnterMain()
        clickTab("我的")
        clickBtn("退出登录")
        waitForText("没有账号？去注册")
        composeRule.onNodeWithText("没有账号？去注册").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  8. 底部导航
    // ════════════════════════════════════════

    @Test
    fun navigation_switchAllTabs() {
        registerAndEnterMain()

        clickTab("发布")
        composeRule.onNodeWithText("发布闲置").assertIsDisplayed()

        clickTab("收藏")
        composeRule.onNodeWithText("我的收藏").assertIsDisplayed()

        clickTab("消息")
        // "消息"出现在 TopBar 标题和 Tab 里，用 hasClickAction 区分
        // 消息页 TopBar 标题不可点击
        composeRule.onNode(hasText("消息") and !hasClickAction()).assertIsDisplayed()

        clickTab("我的")
        // "我的"同理：Tab 可点击，TopBar 标题不可点击
        composeRule.onNode(hasText("我的") and !hasClickAction()).assertIsDisplayed()

        clickTab("首页")
        composeRule.onNodeWithText("校园二手").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  9. 订单流程（下单 → 确认 → 完成）
    // ════════════════════════════════════════

    @Test
    fun order_buyerCreates_sellerConfirms_andCompletes() {
        // ---- 卖家注册并发布商品 ----
        val sellerSid = "seller_$ts"
        registerAndEnterMain(sid = sellerSid, nick = "卖家小王")
        clickTab("发布")
        publishProduct("订单测试商品", "用于测试订单流程", "50")

        // 记住卖家学号，退出
        clickTab("我的")
        clickBtn("退出登录")
        waitForText("没有账号？去注册")

        // ---- 买家注册 ----
        val buyerSid = "buyer_$ts"
        registerAndEnterMain(sid = buyerSid, nick = "买家小李")

        // 首页找到商品 → 进入详情 → 下单
        clickTab("首页")
        waitForText("订单测试商品")
        clickBtn("订单测试商品")
        waitForText("商品详情")

        // 买家下单
        clickBtn("下单购买")
        waitForText("订单已创建")
        composeRule.onNodeWithText("订单已创建").assertIsDisplayed()

        // 买家看订单
        clickTab("我的")
        clickBtn("我的订单")
        waitForText("订单测试商品")
        composeRule.onNodeWithText("待确认").assertIsDisplayed()

        // 买家退出，卖家登录确认订单
        // 先返回个人中心
        clickBtn("我的订单") // 点 TopBar 返回按钮不好定位，直接退出重登
        clickTab("我的")
        clickBtn("退出登录")
        waitForText("没有账号？去注册")

        // ---- 卖家登录 → 确认订单 ----
        composeRule.onNode(hasSetTextAction() and hasText("学号")).performTextInput(sellerSid)
        composeRule.onNode(hasSetTextAction() and hasText("密码")).performTextInput("test1234")
        clickBtn("登录")
        waitForText("校园二手")

        clickTab("我的")
        clickBtn("我的订单")
        waitForText("订单测试商品")
        composeRule.onNodeWithText("待确认").assertIsDisplayed()

        // 卖家确认
        clickBtn("确认订单")
        Espresso.onIdle()
        waitForText("交易中")
        composeRule.onNodeWithText("交易中").assertIsDisplayed()

        // 卖家完成交易
        clickBtn("完成交易")
        Espresso.onIdle()
        waitForText("已完成")
        composeRule.onNodeWithText("已完成").assertIsDisplayed()
    }

    // ════════════════════════════════════════
    //  10. 编辑资料保存
    // ════════════════════════════════════════

    @Test
    fun editProfile_updateNickname_success() {
        registerAndEnterMain(nick = "旧昵称")
        clickTab("我的")
        waitForText("编辑资料")
        clickBtn("编辑资料")

        waitForText("昵称")
        // 清空并输入新昵称（在已有文本后追加标记）
        composeRule.onNode(hasSetTextAction() and hasText("昵称")).performTextInput("新")
        clickBtn("保存")
        waitForText("保存成功")
        composeRule.onNodeWithText("保存成功").assertIsDisplayed()
    }
}