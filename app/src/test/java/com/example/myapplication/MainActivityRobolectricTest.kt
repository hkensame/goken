package com.example.myapplication

import android.os.Bundle
import androidx.lifecycle.Lifecycle
import androidx.test.core.app.ApplicationProvider
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.Robolectric
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.shadows.ShadowLooper

/**
 * Robolectric：测 **Android 组件**（本例 [MainActivity] + [CampusApplication] 上的 Room / 容器初始化），
 * 会执行 `Activity`、`Application.onCreate`、`setContent`、`ComponentActivity`/`enableEdgeToEdge` 等路径，
 * 相比纯 JVM 单测更容易拉高对 **framework / androidx** 相关代码的覆盖率（作业可对比说明）。
 *
 * 与「普通」路径的对照见：`data/session/SessionRepositoryOrdinaryUnitTest`（无 Robolectric）。
 */
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34], application = CampusApplication::class)
class MainActivityRobolectricTest {

    @Test
    fun mainActivity_reachesResumed_andApplicationHasContainer() {
        val controller = Robolectric.buildActivity(MainActivity::class.java)
            .create()
            .start()
            .resume()
        ShadowLooper.idleMainLooper()

        val activity = controller.get()
        assertTrue(activity.lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED))

        val app = activity.application as CampusApplication
        assertNotNull(app.container)
        assertNotNull(app.container.session)
    }

    @Test
    fun mainActivity_hasWindowAfterOnCreate() {
        val activity = Robolectric.buildActivity(MainActivity::class.java)
            .create()
            .get()
        ShadowLooper.idleMainLooper()
        assertNotNull(activity.window)
    }

    @Test
    fun campusApplication_onCreate_runsBeforeActivity() {
        val app = ApplicationProvider.getApplicationContext() as CampusApplication
        assertNotNull(app.container)
        assertNotNull(app.container.db)
    }

    @Test
    fun mainActivity_fullLifecycle_endsDestroyed() {
        val controller = Robolectric.buildActivity(MainActivity::class.java)
            .create()
            .start()
            .resume()
        ShadowLooper.idleMainLooper()

        controller.pause().stop().destroy()

        val activity = controller.get()
        assertEquals(Lifecycle.State.DESTROYED, activity.lifecycle.currentState)
    }

    @Test
    fun mainActivity_recreate_withSavedState_stillResumed() {
        val bundle = Bundle()
        val first = Robolectric.buildActivity(MainActivity::class.java)
            .create()
            .start()
            .resume()
        ShadowLooper.idleMainLooper()
        first.saveInstanceState(bundle)
        first.pause().stop().destroy()

        val second = Robolectric.buildActivity(MainActivity::class.java)
            .create(bundle)
            .start()
            .resume()
        ShadowLooper.idleMainLooper()

        assertTrue(second.get().lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED))
    }
}
