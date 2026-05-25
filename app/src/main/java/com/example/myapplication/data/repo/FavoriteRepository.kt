package com.example.myapplication.data.repo

import com.example.myapplication.data.db.FavoriteDao
import com.example.myapplication.data.db.FavoriteEntity
import kotlinx.coroutines.flow.Flow

class FavoriteRepository(
    private val favoriteDao: FavoriteDao,
) {
    fun observeFavorite(userId: Long, productId: Long): Flow<Boolean> =
        favoriteDao.observeFavorite(userId, productId)

    fun observeFavorites(userId: Long) = favoriteDao.observeFavorites(userId)

    suspend fun toggle(userId: Long, productId: Long) {
        val existing = favoriteDao.find(userId, productId)
        val now = System.currentTimeMillis()
        if (existing != null) {
            favoriteDao.delete(userId, productId)
        } else {
            favoriteDao.insert(FavoriteEntity(userId = userId, productId = productId, createdAt = now))
        }
    }
}
