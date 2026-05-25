package com.example.myapplication.data.repo

import com.example.myapplication.data.db.FavoriteDao
import com.example.myapplication.data.db.FavoriteEntity
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class FavoriteRepositoryTest {
    private val favoriteDao = mockk<FavoriteDao>(relaxed = true)
    private val repo = FavoriteRepository(favoriteDao)

    @Test
    fun toggle_whenExists_deletes(): Unit = runTest {
        coEvery { favoriteDao.find(1L, 200L) } returns FavoriteEntity(
            userId = 1L,
            productId = 200L,
            createdAt = 0L,
        )

        repo.toggle(1L, 200L)

        coVerify(exactly = 1) { favoriteDao.delete(1L, 200L) }
        coVerify(exactly = 0) { favoriteDao.insert(any()) }
    }

    @Test
    fun toggle_whenMissing_inserts(): Unit = runTest {
        coEvery { favoriteDao.find(1L, 200L) } returns null

        val captured = slot<FavoriteEntity>()
        repo.toggle(1L, 200L)

        coVerify(exactly = 1) { favoriteDao.insert(capture(captured)) }
        coVerify(exactly = 0) { favoriteDao.delete(any(), any()) }
        assertEquals(1L, captured.captured.userId)
        assertEquals(200L, captured.captured.productId)
        assertTrue(captured.captured.createdAt > 0L)
    }
}
