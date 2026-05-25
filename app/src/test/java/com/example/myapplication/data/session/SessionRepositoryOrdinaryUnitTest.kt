package com.example.myapplication.data.session

import androidx.datastore.core.DataStore
import androidx.datastore.core.handlers.ReplaceFileCorruptionHandler
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.emptyPreferences
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.take
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.yield
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File

/**
 * **普通 / 原生 JVM 单元测**（无 Robolectric）：`PreferenceDataStoreFactory` + 临时文件 + JaCoCo。
 *
 * 与本包中的 [SessionRepositoryRobolectricTest] **成对存在**，专用于作业里「同一业务、两种测法」对照。
 */
class SessionRepositoryOrdinaryUnitTest {

    @get:Rule
    val tempFolder = TemporaryFolder()

    private lateinit var dataStore: DataStore<Preferences>
    private lateinit var repo: SessionRepository

    @Before
    fun setUp() {
        val root = tempFolder.newFolder("ds")
        dataStore = newPreferencesDataStore(root)
        repo = SessionRepository.forJvmUnitTests(dataStore)
        runBlocking { repo.clear() }
    }

    @After
    fun tearDown() {
        runBlocking { repo.clear() }
    }

    @Test
    fun clear_makesFlowEmitDefaultZero(): Unit = runBlocking {
        repo.setLoggedInUser(77L)
        assertEquals(77L, repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
    }

    @Test
    fun newInstance_readsSameBackingStoreFile(): Unit = runBlocking {
        SessionRepository.forJvmUnitTests(dataStore).apply {
            clear()
            setLoggedInUser(4242L)
        }
        assertEquals(4242L, SessionRepository.forJvmUnitTests(dataStore).userId.first())
    }

    @Test
    fun loginSwitchSequence_endsCleared(): Unit = runBlocking {
        val chain = listOf(101L, 202L, 303L, 404L, 505L)
        chain.forEach { repo.setLoggedInUser(it) }
        assertEquals(chain.last(), repo.userId.first())
        repo.clear()
        assertEquals(0L, repo.userId.first())
    }

    @Test
    fun rapidReassignment_lastWriteWins(): Unit = runBlocking {
        repo.clear()
        repeat(50) { i -> repo.setLoggedInUser(i.toLong()) }
        assertEquals(49L, repo.userId.first())
    }

    @Test
    fun userIdFlow_collectsSequentialEmissions(): Unit = runBlocking {
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

    private fun newPreferencesDataStore(root: File): DataStore<Preferences> =
        PreferenceDataStoreFactory.create(
            corruptionHandler = ReplaceFileCorruptionHandler(produceNewData = { emptyPreferences() }),
            migrations = emptyList(),
            scope = CoroutineScope(Dispatchers.IO + SupervisorJob()),
            produceFile = { File(root, "campus_session_test.preferences_pb") },
        )
}
