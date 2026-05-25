package com.example.myapplication.data

import android.content.Context
import com.example.myapplication.data.db.CampusDatabase
import com.example.myapplication.data.repo.AuthRepository
import com.example.myapplication.data.repo.ChatRepository
import com.example.myapplication.data.repo.FavoriteRepository
import com.example.myapplication.data.repo.OrderRepository
import com.example.myapplication.data.repo.ProductRepository
import com.example.myapplication.data.session.SessionRepository

class AppContainer(
    context: Context,
    val db: CampusDatabase,
) {
    private val appContext = context.applicationContext

    val session = SessionRepository(appContext)
    val auth = AuthRepository(db.userDao(), session)
    val products = ProductRepository(db.productDao(), appContext)
    val favorites = FavoriteRepository(db.favoriteDao())
    val orders = OrderRepository(db.orderDao(), db.productDao())
    val chat = ChatRepository(db)
}
