package com.example.myapplication.data.repo

import android.content.Context
import android.net.Uri
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.data.db.ProductEntity
import com.example.myapplication.domain.ProductStatus
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.io.ByteArrayInputStream
import java.io.File

class ProductRepositoryTest {
    private lateinit var filesDir: File
    private val productDao = mockk<ProductDao>()
    private val context = mockk<Context>()
    private lateinit var repo: ProductRepository

    @Before
    fun setUp() {
        filesDir = File.createTempFile("product_repo", "").apply {
            delete()
            mkdirs()
        }
        every { context.applicationContext } returns context
        every { context.filesDir } returns filesDir
        repo = ProductRepository(productDao, context)
    }

    @After
    fun tearDown() {
        filesDir.deleteRecursively()
    }

    @Test
    fun publish_invalidPrice_fails(): Unit = runTest {
        val r = repo.publish(1L, "t", "d", "12abc", "其他", null)
        assertTrue(r.isFailure)
    }

    @Test
    fun publish_negativePrice_fails(): Unit = runTest {
        val r = repo.publish(1L, "t", "d", "-1", "其他", null)
        assertTrue(r.isFailure)
    }

    @Test
    fun publish_withoutImage_insertsOnSaleAndTrimsText(): Unit = runTest {
        coEvery { productDao.insert(any()) } returns 55L

        val r = repo.publish(
            sellerId = 3L,
            title = "  desk  ",
            description = " ok ",
            priceYuan = "12.5",
            category = "日用品",
            imageUri = null,
        )
        assertTrue(r.isSuccess)
        assertEquals(55L, r.getOrNull())

        val captured = slot<ProductEntity>()
        coVerify(exactly = 1) { productDao.insert(capture(captured)) }
        assertEquals("desk", captured.captured.title)
        assertEquals("ok", captured.captured.description)
        assertEquals(1250L, captured.captured.priceCents)
        assertEquals(ProductStatus.ON_SALE, captured.captured.status)
        assertEquals(null, captured.captured.imageLocalPath)
    }

    @Test
    fun publish_withImage_copiesToInternalStorage(): Unit = runTest {
        val uri = mockk<Uri>()
        val resolver = mockk<android.content.ContentResolver>()
        every { context.contentResolver } returns resolver
        every { resolver.openInputStream(uri) } returns ByteArrayInputStream(byteArrayOf(1, 2, 3))
        coEvery { productDao.insert(any()) } returns 9L

        val r = repo.publish(
            sellerId = 1L,
            title = "pic",
            description = "d",
            priceYuan = "1",
            category = "其他",
            imageUri = uri,
        )
        assertTrue(r.isSuccess)

        val captured = slot<ProductEntity>()
        coVerify(exactly = 1) { productDao.insert(capture(captured)) }
        val path = captured.captured.imageLocalPath
        assertTrue(path != null && File(path).exists())
    }
}
