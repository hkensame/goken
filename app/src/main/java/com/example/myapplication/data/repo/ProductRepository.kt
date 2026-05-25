package com.example.myapplication.data.repo

import android.content.Context
import android.net.Uri
import com.example.myapplication.data.db.ProductDao
import com.example.myapplication.data.db.ProductEntity
import com.example.myapplication.domain.ProductStatus
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext
import java.io.File
import java.util.UUID

class ProductRepository(
    private val productDao: ProductDao,
    private val appContext: Context,
) {
    fun observeMarket(
        search: String,
        category: String?,
        categoryAll: Boolean,
    ): Flow<List<com.example.myapplication.data.db.ProductListItem>> =
        productDao.observeProducts(
            q = search.trim(),
            category = category ?: "",
            categoryAll = if (categoryAll) 1 else 0,
            status = ProductStatus.ON_SALE,
        )

    fun observeProduct(id: Long) = productDao.observeById(id)

    suspend fun getProduct(id: Long) = productDao.getById(id)

    suspend fun publish(
        sellerId: Long,
        title: String,
        description: String,
        priceYuan: String,
        category: String,
        imageUri: Uri?,
    ): Result<Long> = withContext(Dispatchers.IO) {
        val yuan = priceYuan.toDoubleOrNull()
            ?: return@withContext Result.failure(IllegalArgumentException("价格格式不正确"))
        if (yuan < 0) return@withContext Result.failure(IllegalArgumentException("价格不能为负"))
        val path = imageUri?.let { copyImageToInternal(it) }
        val now = System.currentTimeMillis()
        val id = productDao.insert(
            ProductEntity(
                sellerId = sellerId,
                title = title.trim(),
                description = description.trim(),
                priceCents = (yuan * 100).toLong(),
                category = category,
                imageLocalPath = path,
                status = ProductStatus.ON_SALE,
                createdAt = now,
            ),
        )
        Result.success(id)
    }

    private fun copyImageToInternal(uri: Uri): String {
        val dir = File(appContext.filesDir, "product_images").apply { mkdirs() }
        val out = File(dir, "${UUID.randomUUID()}.jpg")
        appContext.contentResolver.openInputStream(uri)?.use { input ->
            out.outputStream().use { output -> input.copyTo(output) }
        } ?: throw IllegalArgumentException("无法读取图片")
        return out.absolutePath
    }
}

