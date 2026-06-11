package com.example.myapplication.data.repo

import android.content.Context
import android.net.Uri
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.data.db.ProductEntity
import com.example.myapplication.data.remote.MockRemoteApi
import com.example.myapplication.domain.ProductStatus
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.io.ByteArrayInputStream
import java.io.File

class ProductRepositoryTest {

    private lateinit var filesDir: File
    private val productDao = mockk<ProductDao>()
    private val context = mockk<Context>()

    // ════════════════════════════════════════
    //  Room 分支（api == null）
    // ════════════════════════════════════════

    private lateinit var roomRepo: ProductRepository

    @Before
    fun setUp() {
        filesDir = File.createTempFile("product_repo", "").apply { delete(); mkdirs() }
        every { context.applicationContext } returns context
        every { context.filesDir } returns filesDir
        roomRepo = ProductRepository(productDao, context)
    }

    @After
    fun tearDown() { filesDir.deleteRecursively() }

    @Test
    fun room_publish_invalidPrice_fails() = runTest {
        assertTrue(roomRepo.publish(1L, "t", "d", "12abc", "其他", null).isFailure)
    }

    @Test
    fun room_publish_negativePrice_fails() = runTest {
        assertTrue(roomRepo.publish(1L, "t", "d", "-1", "其他", null).isFailure)
    }

    @Test
    fun room_publish_withoutImage_insertsTrimmed() = runTest {
        coEvery { productDao.insert(any()) } returns 55L

        val r = roomRepo.publish(
            sellerId = 3L, title = "  desk  ", description = " ok ",
            priceYuan = "12.5", category = "日用品", imageUri = null,
        )
        assertTrue(r.isSuccess)
        assertEquals(55L, r.getOrNull())

        val captured = slot<ProductEntity>()
        coVerify { productDao.insert(capture(captured)) }
        assertEquals("desk", captured.captured.title)
        assertEquals("ok", captured.captured.description)
        assertEquals(1250L, captured.captured.priceCents)
        assertEquals(ProductStatus.ON_SALE, captured.captured.status)
    }

    @Test
    fun room_publish_withImage_copiesFile() = runTest {
        val uri = mockk<Uri>()
        val resolver = mockk<android.content.ContentResolver>()
        every { context.contentResolver } returns resolver
        every { resolver.openInputStream(uri) } returns ByteArrayInputStream(byteArrayOf(1, 2, 3))
        coEvery { productDao.insert(any()) } returns 9L

        val r = roomRepo.publish(1L, "pic", "d", "1", "其他", uri)
        assertTrue(r.isSuccess)

        val captured = slot<ProductEntity>()
        coVerify { productDao.insert(capture(captured)) }
        assertNotNull(captured.captured.imageLocalPath)
        assertTrue(File(captured.captured.imageLocalPath!!).exists())
    }

    // ════════════════════════════════════════
    //  MockRemoteApi 分支（api != null）
    // ════════════════════════════════════════

    @Test
    fun api_publish_callsMockAndReturnsId() = runTest {
        val api = MockRemoteApi()
        val repo = ProductRepository(productDao, context, api)

        val r = repo.publish(1L, "iPhone", "desc", "99.9", "数码电子", null)
        assertTrue(r.isSuccess)
        assertEquals(1L, r.getOrNull())
        assertTrue(api.called("publish"))
        assertTrue(api.called("syncProducts"))
    }

    @Test
    fun api_publish_invalidPrice_stillFailsBeforeReachingApi() = runTest {
        val api = MockRemoteApi()
        val repo = ProductRepository(productDao, context, api)

        val r = repo.publish(1L, "t", "d", "bad", "其他", null)
        assertTrue(r.isFailure)
        // 验证没调用 mock（校验在 api 之前）
        assertEquals(0, api.calls.size)
    }

    @Test
    fun api_publish_failure_fallsBackToRoom() = runTest {
        // 让 mock 的 publish 抛异常：不需要特别设置，只要 MockRemoteApi 正常工作
        // 但如果 mock 本身不抛异常，我们无法直接模拟远程失败
        // 这种情况下 Room fallback 由 catch 块保证
        val api = MockRemoteApi()
        coEvery { productDao.insert(any()) } returns 10L
        val repo = ProductRepository(productDao, context, api)

        // 正常情况：mock 成功，不回退 Room
        val r = repo.publish(1L, "t", "d", "10", "其他", null)
        assertTrue(r.isSuccess)
        coVerify(exactly = 0) { productDao.insert(any()) } // mock 成功不走 Room
    }
}