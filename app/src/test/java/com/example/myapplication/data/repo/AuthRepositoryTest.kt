package com.example.myapplication.data.repo

import com.example.myapplication.data.db.UserDao
import com.example.myapplication.data.db.UserEntity
import com.example.myapplication.data.session.SessionRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * MockK：把 [UserDao]、[SessionRepository] 换成可控替身，只测 [AuthRepository] 的分支逻辑。
 */
class AuthRepositoryTest {
    private val userDao = mockk<UserDao>()
    private val session = mockk<SessionRepository>(relaxUnitFun = true)

    private val repo = AuthRepository(userDao, session)

    @Test
    fun register_blankStudentId_fails(): Unit = runTest {
        val r = repo.register(" ", "pw", "nick")
        assertTrue(r.isFailure)
    }

    @Test
    fun register_blankPassword_fails(): Unit = runTest {
        assertTrue(repo.register("2024100", " ", "nick").isFailure)
    }

    @Test
    fun register_trimsStudentIdOnInsert(): Unit = runTest {
        coEvery { userDao.getByStudentId("  S1  ") } returns null
        val capturedUser = slot<UserEntity>()
        coEvery { userDao.insert(capture(capturedUser)) } returns 99L
        val r = repo.register("  S1  ", "pw", "n")
        assertTrue(r.isSuccess)
        assertEquals("S1", capturedUser.captured.studentId)
    }

    @Test
    fun logout_callsSessionClear(): Unit = runTest {
        repo.logout()
        coVerify(exactly = 1) { session.clear() }
    }

    @Test
    fun register_duplicateStudentId_fails(): Unit = runTest {
        coEvery { userDao.getByStudentId("2024001") } returns UserEntity(
            id = 1L,
            studentId = "2024001",
            password = "x",
            nickname = "a",
            createdAt = 0L,
        )
        val r = repo.register("2024001", "pw", "nick")
        assertTrue(r.isFailure)
    }

    @Test
    fun register_insertsAndSetsSession(): Unit = runTest {
        coEvery { userDao.getByStudentId("2024002") } returns null
        coEvery { userDao.insert(any()) } returns 42L

        val r = repo.register("2024002", "secret", "小李")
        assertTrue(r.isSuccess)
        assertEquals(42L, r.getOrNull())
        coVerify(exactly = 1) { session.setLoggedInUser(42L) }
    }

    @Test
    fun login_trimsStudentId_beforeLookup(): Unit = runTest {
        val user = UserEntity(id = 1L, studentId = "U1", password = "pw", nickname = "n", createdAt = 0L)
        coEvery { userDao.login("U1", "pw") } returns user
        val r = repo.login("  U1  ", "pw")
        assertTrue(r.isSuccess)
        assertEquals(1L, r.getOrNull())
    }

    @Test
    fun login_unknownUser_fails(): Unit = runTest {
        coEvery { userDao.login("2024003", "pw") } returns null
        val r = repo.login("2024003", "pw")
        assertTrue(r.isFailure)
    }

    @Test
    fun login_success_setsSession(): Unit = runTest {
        val user = UserEntity(
            id = 7L,
            studentId = "2024004",
            password = "pw",
            nickname = "小王",
            createdAt = 1L,
        )
        coEvery { userDao.login("2024004", "pw") } returns user

        val r = repo.login("2024004", "pw")
        assertTrue(r.isSuccess)
        assertEquals(7L, r.getOrNull())
        coVerify(exactly = 1) { session.setLoggedInUser(7L) }
    }
}
