package com.example.myapplication.data.session

import android.app.Application
import android.content.Context
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.take
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.yield
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * 仪器测试（真机 / 模拟器）：走**原生 Android + 真实 DataStore I/O**，用于和 Robolectric 路径做覆盖率对比。
 *
 * 测试逻辑与 [SessionRepositoryRobolectricTest] 尽量镜像，便于作业里对比同一代码在两套环境下的覆盖差异。
 */
@RunWith(AndroidJUnit4::class)
class SessionRepositoryInstrumentedTest {

    private lateinit var app: Application

    private fun targetContext(): Context =
        InstrumentationRegistry.getInstrumentation().targetContext

    @Before
    fun setUp() {
        app = targetContext().applicationContext as Application
        runBlocking { SessionRepository(app).clear() }
    }

    @After
    fun tearDown() {
        runBlocking { SessionRepository(app).clear() }
    }

    @Test
    fun clear_makesFlowEmitDefaultZero(): Unit = runBlocking {
        val repo = SessionRepository(app)
        repo.setLoggedInUser(77L)
        assertEquals(77L, repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
    }

    @Test
    fun newInstance_readsDataStoreBackedBySameApplication(): Unit = runBlocking {
        SessionRepository(app).apply {
            clear()
            setLoggedInUser(4242L)
        }
        assertEquals(4242L, SessionRepository(app).userId.first())
    }

    @Test
    fun loginSwitchSequence_endsCleared(): Unit = runBlocking {
        val repo = SessionRepository(app)
        val chain = listOf(101L, 202L, 303L, 404L, 505L)
        chain.forEach { repo.setLoggedInUser(it) }
        assertEquals(chain.last(), repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
    }

    @Test
    fun rapidReassignment_lastWriteWins(): Unit = runBlocking {
        val repo = SessionRepository(app)
        repo.clear()
        repeat(50) { i -> repo.setLoggedInUser(i.toLong()) }
        assertEquals(49L, repo.userId.first())
    }

    @Test
    fun userIdFlow_collectsSequentialEmissions(): Unit = runBlocking {
        val repo = SessionRepository(app)
        repo.clear()

        val history = async {
            repo.userId.take(4).toList()
        }
        yield()
        repo.setLoggedInUser(11L)
        repo.setLoggedInUser(22L)
        repo.setLoggedInUser(33L)

        assertEquals(listOf(0L, 11L, 22L, 33L), history.await())
    }

    @Test
    fun toggleTwiceBetweenUserAndAnonymous_behaves(): Unit = runBlocking {
        val repo = SessionRepository(app)
        repo.clear()
        repo.setLoggedInUser(900L)
        assertEquals(900L, repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
        repo.setLoggedInUser(901L)
        assertEquals(901L, repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
    }
}
