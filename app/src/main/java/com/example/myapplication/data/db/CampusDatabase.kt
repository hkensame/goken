package com.example.myapplication.data.db

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase

@Database(
    entities = [
        UserEntity::class,
        ProductEntity::class,
        FavoriteEntity::class,
        OrderEntity::class,
        ConversationEntity::class,
        MessageEntity::class,
    ],
    version = 1,
    exportSchema = false,
)
abstract class CampusDatabase : RoomDatabase() {
    abstract fun userDao(): UserDao
    abstract fun productDao(): ProductDao
    abstract fun favoriteDao(): FavoriteDao
    abstract fun orderDao(): OrderDao
    abstract fun conversationDao(): ConversationDao
    abstract fun messageDao(): MessageDao

    companion object {
        fun build(context: Context): CampusDatabase =
            Room.databaseBuilder(
                context.applicationContext,
                CampusDatabase::class.java,
                "campus_trade.db",
            ).fallbackToDestructiveMigration()
                .build()
    }
}
