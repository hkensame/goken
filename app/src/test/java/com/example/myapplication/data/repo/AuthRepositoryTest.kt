package com.example.myapplication.data.repo

import com.example.myapplication.data.db.UserDao
import com.example.myapplication.data.db.UserEntity
import com.example.myapplication.data.remote.MockRemoteApi
import com.example.myapplication.data.session.SessionRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class AuthRepositoryTest {

    // ════════════════════════════════════════
    //  Room 分支（api == null）
    // ════════════════════════════════════════

    private val userDao = mockk<UserDao>(relaxUnitFun = true)
    private val session = mockk<SessionRepository>(relaxUnitFun = true)
    private val roomRepo = AuthRepository(userDao, session)

    @Test
    fun room_register_blankStudentId_fails() = runTest {
        assertTrue(roomRepo.register(" ", "pw", "nick").isFailure)
    }

    @Test
    fun room_register_blankPassword_fails() = runTest {
        assertTrue(roomRepo.register("S1", " ", "nick").isFailure)
    }

    @Test
    fun room_register_trimsStudentId() = runTest {
        coEvery { userDao.getByStudentId("S1") } returns null
        val captured = slot<UserEntity>()
        coEvery { userDao.insert(capture(captured)) } returns 99L

        val r = roomRepo.register("  S1  ", "pw", "n")
        assertTrue(r.isSuccess)
        assertEquals("S1", captured.captured.studentId)
    }

    @Test
    fun room_register_duplicate_fails() = runTest {
        coEvery { userDao.getByStudentId("S1") } returns UserEntity(
            id = 1L, studentId = "S1", password = "x", nickname = "a", createdAt = 0L
        )
        assertTrue(roomRepo.register("S1", "pw", "nick").isFailure)
    }

    @Test
    fun room_register_success_setsSession() = runTest {
        coEvery { userDao.getByStudentId("S1") } returns null
        coEvery { userDao.insert(any()) } returns 42L

        val r = roomRepo.register("S1", "pw", "小李")
        assertTrue(r.isSuccess)
        assertEquals(42L, r.getOrNull())
        coVerify { session.setLoggedInUser(42L) }
    }

    @Test
    fun room_login_wrongCredentials_fails() = runTest {
        coEvery { userDao.login("S1", "pw") } returns null
        assertTrue(roomRepo.login("S1", "pw").isFailure)
    }

    @Test
    fun room_login_success_setsSession() = runTest {
        val user = UserEntity(id = 7L, studentId = "S1", password = "pw", nickname = "n", createdAt = 0L)
        coEvery { userDao.login("S1", "pw") } returns user

        val r = roomRepo.login("S1", "pw")
        assertTrue(r.isSuccess)
        assertEquals(7L, r.getOrNull())
        coVerify { session.setLoggedInUser(7L) }
    }

    @Test
    fun room_logout_clearsSession() = runTest {
        roomRepo.logout()
        coVerify { session.clear() }
    }

    @Test
    fun room_updateProfile_callsDao() = runTest {
        roomRepo.updateProfile(1L, "newName", "13800001111")
        coVerify { userDao.updateNickname(1L, "newName", any()) }
        coVerify { userDao.updatePhone(1L, "13800001111", any()) }

    }

    @Test
    fun room_updateAvatar_callsDao() = runTest {
        roomRepo.updateAvatar(1L, "/img/a.jpg")
        coVerify { userDao.updateAvatar(1L, "/img/a.jpg", any()) }
    }

    // ════════════════════════════════════════
    //  MockRemoteApi 分支（api != null）
    // ════════════════════════════════════════

    @Test
    fun api_register_callsMockAndSetsSession() = runTest {
        val api = MockRemoteApi()
        val repo = AuthRepository(userDao, session, api)

        val r = repo.register("S1", "pw", "小李")
        assertTrue(r.isSuccess)
        assertEquals(1L, r.getOrNull())
        coVerify { session.setLoggedInUser(1L) }
        assertTrue(api.called("register"))
        assertTrue(api.called("syncUsers"))
    }

    @Test
    fun api_register_duplicate_fallsBackToRoom() = runTest {
        val api = MockRemoteApi()
        api.register("S1", "pw", "A", 0L) // 先注册一个，让 mock 抛异常

        // Room fallback 也检测重复
        coEvery { userDao.getByStudentId("S1") } returns UserEntity(
            id = 1L, studentId = "S1", password = "x", nickname = "a", createdAt = 0L
        )
        val repo = AuthRepository(userDao, session, api)
        val r = repo.register("S1", "pw2", "B")
        assertTrue(r.isFailure)
    }

    @Test
    fun api_updateProfile_callsMockAndRoom() = runTest {
        val api = MockRemoteApi()
        api.register("S1", "pw", "A", 0L)
        val repo = AuthRepository(userDao, session, api)

        repo.updateProfile(1L, "newName", "13800001111")
        assertTrue(api.called("updateProfile"))
        coVerify { userDao.updateNickname(1L, "newName", any()) }
    }

    @Test
    fun api_updateAvatar_callsMockAndRoom() = runTest {
        val api = MockRemoteApi()
        api.register("S1", "pw", "A", 0L)
        val repo = AuthRepository(userDao, session, api)

        repo.updateAvatar(1L, "/img/a.jpg")
        assertTrue(api.called("updateAvatar"))
        coVerify { userDao.updateAvatar(1L, "/img/a.jpg", any()) }
    }
}