package com.example.myapplication.domain

import org.junit.Assert.assertTrue
import org.junit.Test

class ProductCategoriesTest {
    @Test
    fun all_isNonEmptyAndContainsBookCategory() {
        assertTrue(ProductCategories.all.isNotEmpty())
        assertTrue(ProductCategories.all.any { it.contains("教材") })
    }
}
