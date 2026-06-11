package com.example.myapplication.data.repo

import com.example.myapplication.data.db.UserDao
import com.example.myapplication.data.db.UserEntity
import com.example.myapplication.data.remote.RemoteApi
import com.example.myapplication.data.session.SessionRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AuthRepository(
    private val userDao: UserDao,
    private val session: SessionRepository,
    private val api: RemoteApi? = null,
) {
    constructor(userDao: UserDao, session: SessionRepository) : this(userDao, session, null)

    suspend fun register(studentId: String, password: String, nickname: String): Result<Long> {
        if (studentId.isBlank() || password.isBlank()) {
            return Result.failure(IllegalArgumentException("学号/密码不能为空"))
        }
        val now = System.currentTimeMillis()
        val trimmed = studentId.trim()
        val nick = nickname.ifBlank { trimmed }

        if (api != null) {
            try {
                val id = withContext(Dispatchers.IO) { api.register(trimmed, password, nick, now) }
                api.syncUsers()
                session.setLoggedInUser(id)
                return Result.success(id)
            } catch (_: Exception) { /* 远程失败，回退 Room */ }
        }

        if (userDao.getByStudentId(trimmed) != null) {
            return Result.failure(IllegalStateException("该学号已注册"))
        }
        val id = userDao.insert(UserEntity(studentId = trimmed, password = password, nickname = nick, createdAt = now))
        session.setLoggedInUser(id)
        return Result.success(id)
    }

    suspend fun login(studentId: String, password: String): Result<Long> {
        val trimmed = studentId.trim()
        // 远程可用时先同步
        if (api != null) {
            try { api.syncUsers() } catch (_: Exception) {}
        }
        val user = userDao.login(trimmed, password)
            ?: return Result.failure(IllegalArgumentException("学号或密码错误"))
        session.setLoggedInUser(user.id)
        return Result.success(user.id)
    }

    suspend fun logout() = session.clear()

    suspend fun user(id: Long) = userDao.getById(id)

    suspend fun updateProfile(id: Long, nickname: String, phone: String?) {
        val now = System.currentTimeMillis()
        try { api?.updateProfile(id, nickname, phone, now) } catch (_: Exception) {}
        userDao.updateNickname(id, nickname, now)
        userDao.updatePhone(id, phone, now)
    }

    suspend fun updateAvatar(id: Long, avatarPath: String?) {
        val now = System.currentTimeMillis()
        try { api?.updateAvatar(id, avatarPath, now) } catch (_: Exception) {}
        userDao.updateAvatar(id, avatarPath, now)
    }
}
