package com.example.myapplication.data.repo

import com.example.myapplication.data.db.FavoriteDao
import com.example.myapplication.data.db.FavoriteEntity
import com.example.myapplication.data.remote.MockRemoteApi
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class FavoriteRepositoryTest {

    // ════════════════════════════════════════
    //  Room 分支（api == null）
    // ════════════════════════════════════════

    private val favoriteDao = mockk<FavoriteDao>(relaxed = true)
    private val roomRepo = FavoriteRepository(favoriteDao)

    @Test
    fun room_toggle_whenExists_deletes() = runTest {
        coEvery { favoriteDao.find(1L, 200L) } returns FavoriteEntity(
            userId = 1L, productId = 200L, createdAt = 0L,
        )

        roomRepo.toggle(1L, 200L)

        coVerify(exactly = 1) { favoriteDao.delete(1L, 200L) }
        coVerify(exactly = 0) { favoriteDao.insert(any()) }
    }

    @Test
    fun room_toggle_whenMissing_inserts() = runTest {
        coEvery { favoriteDao.find(1L, 200L) } returns null

        roomRepo.toggle(1L, 200L)

        val captured = slot<FavoriteEntity>()
        coVerify { favoriteDao.insert(capture(captured)) }
        coVerify(exactly = 0) { favoriteDao.delete(any(), any()) }
        assertEquals(1L, captured.captured.userId)
        assertEquals(200L, captured.captured.productId)
        assertTrue(captured.captured.createdAt > 0L)
    }

    // ════════════════════════════════════════
    //  MockRemoteApi 分支（api != null）
    // ════════════════════════════════════════

    @Test
    fun api_toggle_whenMissing_callsAddFavorite() = runTest {
        val api = MockRemoteApi()
        coEvery { favoriteDao.find(1L, 10L) } returns null
        val repo = FavoriteRepository(favoriteDao, api)

        repo.toggle(1L, 10L)

        assertTrue(api.called("addFavorite"))
        assertTrue(api.called("syncFavorites"))
        // mock 成功时不再走 Room
        coVerify(exactly = 0) { favoriteDao.insert(any()) }
    }

    @Test
    fun api_toggle_whenExists_callsRemoveFavorite() = runTest {
        val api = MockRemoteApi()
        coEvery { favoriteDao.find(1L, 10L) } returns FavoriteEntity(
            userId = 1L, productId = 10L, createdAt = 0L,
        )
        val repo = FavoriteRepository(favoriteDao, api)

        repo.toggle(1L, 10L)

        assertTrue(api.called("removeFavorite"))
        // mock 成功时不再走 Room
        coVerify(exactly = 0) { favoriteDao.delete(any(), any()) }
    }
}