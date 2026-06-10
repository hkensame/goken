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
 * Espresso 体系下的 **仪器化 UI 集成测试**（真机 / 模拟器）。
 *
 * 本 App 使用 Jetpack Compose，因此通过 [createAndroidComposeRule] 启动 [MainActivity]，
 * 用 Compose Test API 查找/操作节点；底层仍由 Espresso 负责主线程同步（[Espresso.onIdle]）。
 *
 * 与 Robolectric 对照：[MainActivityRobolectricTest]（`src/test`）；
 * 与 DataStore 仪器测对照：[data.session.SessionRepositoryInstrumentedTest]。
 */
@RunWith(AndroidJUnit4::class)
class LoginFlowEspressoTest {

    @get:Rule
    val composeRule = createAndroidComposeRule<MainActivity>()

    private fun app(): CampusApplication =
        InstrumentationRegistry.getInstrumentation()
            .targetContext
            .applicationContext as CampusApplication

    @Before
    fun ensureLoggedOut(): Unit = runBlocking {
        app().container.auth.logout()
    }

    @After
    fun tearDown(): Unit = runBlocking {
        app().container.auth.logout()
    }



    @Test
    fun loginScreen_navigateToRegisterAndBack() {
        composeRule.onNodeWithText("没有账号？去注册").performClick()
        Espresso.onIdle()

        composeRule.onNodeWithText("注册").assertIsDisplayed()
        composeRule.onNodeWithText("昵称（可选）").assertIsDisplayed()

        composeRule.onNodeWithText("返回登录").performClick()
        Espresso.onIdle()

        composeRule.onNodeWithText("没有账号？去注册").assertIsDisplayed()
    }

    @Test
    fun registerAndLogin_entersMainShell() {
        val studentId = "espresso_${System.currentTimeMillis()}"
        val password = "test-pass"

        composeRule.onNodeWithText("没有账号？去注册").performClick()
        Espresso.onIdle()

        fillCredentials(studentId = studentId, password = password, nickname = "Espresso用户")

        composeRule.onNode(hasText("注册并登录") and hasClickAction()).performClick()
        Espresso.onIdle()

        composeRule.waitUntil(timeoutMillis = 10_000) {
            composeRule.onAllNodes(hasText("校园二手")).fetchSemanticsNodes().isNotEmpty()
        }
        composeRule.onNodeWithText("校园二手").assertIsDisplayed()
        composeRule.onNodeWithText("首页").assertIsDisplayed()
        composeRule.onNodeWithText("发布").assertIsDisplayed()
    }

    @Test
    fun loginWithWrongPassword_showsError() {
        val studentId = "espresso_err_${System.currentTimeMillis()}"
        val password = "correct-pass"

        registerUser(studentId, password)

        composeRule.onNode(hasText("登录") and hasClickAction()).performClick()
        Espresso.onIdle()

        composeRule.onNodeWithText("学号或密码错误").assertIsDisplayed()
    }

    private fun registerUser(studentId: String, password: String) {
        composeRule.onNodeWithText("没有账号？去注册").performClick()
        Espresso.onIdle()
        fillCredentials(studentId, password, nickname = "")
        composeRule.onNode(hasText("注册并登录") and hasClickAction()).performClick()
        Espresso.onIdle()
        composeRule.waitUntil(timeoutMillis = 10_000) {
            composeRule.onAllNodes(hasText("校园二手")).fetchSemanticsNodes().isNotEmpty()
        }
        runBlocking { app().container.auth.logout() }
        Espresso.onIdle()
        composeRule.waitUntil(timeoutMillis = 10_000) {
            composeRule.onAllNodes(hasText("登录")).fetchSemanticsNodes().isNotEmpty()
        }
        fillCredentials(studentId, "wrong-password", nickname = "")
    }

    private fun fillCredentials(studentId: String, password: String, nickname: String) {
        val fields = composeRule.onAllNodes(hasSetTextAction())
        fields[0].performTextInput(studentId)
        fields[1].performTextInput(password)
        if (nickname.isNotEmpty() && fields.fetchSemanticsNodes().size > 2) {
            fields[2].performTextInput(nickname)
        }
    }
}
