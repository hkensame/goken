package com.example.myapplication.data.session

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.dataStore by preferencesDataStore(name = "campus_session")

class SessionRepository private constructor(
    private val ds: DataStore<Preferences>,
) {
    /** App 与仪器测：走 [Context] 上的单例 DataStore（与单测注入的 store 行为一致）。 */
    constructor(context: Context) : this(context.applicationContext.dataStore)

    companion object {
        /**
         * 供 [SessionRepositoryOrdinaryUnitTest] 使用：普通 JVM 单测注入 [DataStore]（如 `PreferenceDataStoreFactory` + 临时文件）。
         * 与 Robolectric 路径对照时，体现无影子 Android、手写 DataStore 的写法。
         */
        fun forJvmUnitTests(dataStore: DataStore<Preferences>): SessionRepository =
            SessionRepository(dataStore)
    }

    private val keyUserId = longPreferencesKey("user_id")

    /** 未登录时为 0 */
    val userId: Flow<Long> = ds.data.map { prefs ->
        prefs[keyUserId] ?: 0L
    }

    suspend fun setLoggedInUser(id: Long) {
        ds.edit { it[keyUserId] = id }
    }

    suspend fun clear() {
        ds.edit { it.remove(keyUserId) }
    }
}
