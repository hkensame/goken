package com.example.myapplication.data.repo

import com.example.myapplication.data.db.FavoriteDao
import com.example.myapplication.data.db.FavoriteEntity
import com.example.myapplication.data.remote.RemoteApi
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext

class FavoriteRepository(
    private val favoriteDao: FavoriteDao,
    private val api: RemoteApi? = null,
) {
    constructor(favoriteDao: FavoriteDao) : this(favoriteDao, null)

    fun observeFavorite(userId: Long, productId: Long): Flow<Boolean> =
        favoriteDao.observeFavorite(userId, productId)

    fun observeFavorites(userId: Long) = favoriteDao.observeFavorites(userId)

    suspend fun toggle(userId: Long, productId: Long) {
        val existing = favoriteDao.find(userId, productId)
        val now = System.currentTimeMillis()

        if (api != null) {
            try {
                withContext(Dispatchers.IO) {
                    if (existing != null) api.removeFavorite(userId, productId)
                    else api.addFavorite(userId, productId, now)
                    api.syncFavorites(userId)
                }
                return
            } catch (_: Exception) { /* 远程失败，回退 Room */ }
        }

        if (existing != null) {
            favoriteDao.delete(userId, productId)
        } else {
            favoriteDao.insert(FavoriteEntity(userId = userId, productId = productId, createdAt = now))
        }
    }
}
