package com.example.myapplication.data.repo

import com.example.myapplication.data.db.UserDao
import com.example.myapplication.data.db.UserEntity
import com.example.myapplication.data.session.SessionRepository

class AuthRepository(
    private val userDao: UserDao,
    private val session: SessionRepository,
) {
    suspend fun register(studentId: String, password: String, nickname: String): Result<Long> {
        if (studentId.isBlank() || password.isBlank()) {
            return Result.failure(IllegalArgumentException("学号/密码不能为空"))
        }
        if (userDao.getByStudentId(studentId) != null) {
            return Result.failure(IllegalStateException("该学号已注册"))
        }
        val now = System.currentTimeMillis()
        val id = userDao.insert(
            UserEntity(
                studentId = studentId.trim(),
                password = password,
                nickname = nickname.ifBlank { studentId },
                createdAt = now,
            ),
        )
        session.setLoggedInUser(id)
        return Result.success(id)
    }

    suspend fun login(studentId: String, password: String): Result<Long> {
        val user = userDao.login(studentId.trim(), password)
            ?: return Result.failure(IllegalArgumentException("学号或密码错误"))
        session.setLoggedInUser(user.id)
        return Result.success(user.id)
    }

    suspend fun logout() = session.clear()

    suspend fun user(id: Long) = userDao.getById(id)
}
