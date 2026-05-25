plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.google.ksp)
    id("jacoco")
}

android {
    namespace = "com.example.myapplication"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.example.myapplication"
        minSdk = 24
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        debug {
            enableUnitTestCoverage = true
            enableAndroidTestCoverage = true
        }
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    testOptions {
        unitTests {
            isIncludeAndroidResources = true
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_21
        targetCompatibility = JavaVersion.VERSION_21
    }

    kotlinOptions {
        jvmTarget = "21"
    }
    buildFeatures {
        compose = true
    }
}



dependencies {
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.runtime)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material3.adaptive.navigation.suite)
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.navigation.compose)
    implementation(libs.androidx.room.runtime)
    implementation(libs.androidx.room.ktx)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.coil.compose)
    implementation(libs.androidx.compose.material.icons.extended)
    implementation(libs.core.ktx)
    ksp(libs.androidx.room.compiler)
    testImplementation(libs.junit)
    testImplementation(libs.androidx.test.core)
    testImplementation(libs.mockk)
    testImplementation("org.robolectric:robolectric:4.11.1")
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.robolectric)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.compose.ui.test.junit4)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.test.core)
    androidTestImplementation(libs.kotlinx.coroutines.test)
    debugImplementation(libs.androidx.compose.ui.test.manifest)
    debugImplementation(libs.androidx.compose.ui.tooling)
}

jacoco {
    toolVersion = libs.versions.jacoco.get()
}

tasks.withType<Test>().configureEach {
    configure<org.gradle.testing.jacoco.plugins.JacocoTaskExtension> {
        isIncludeNoLocationClasses = true
        excludes = listOf("jdk.internal.*")
    }
}

tasks.register<org.gradle.testing.jacoco.tasks.JacocoReport>("jacocoTestReport") {
    group = "verification"
    description = "JaCoCo HTML/XML for :app debug unit tests (compare totals over time)."
    val testTask = tasks.named<Test>("testDebugUnitTest")
    dependsOn(testTask)

    val jacocoExt = testTask.get().extensions.getByType(org.gradle.testing.jacoco.plugins.JacocoTaskExtension::class.java)
    executionData.setFrom(jacocoExt.destinationFile)

    reports {
        xml.required.set(true)
        html.required.set(true)
    }

    val classFilter = listOf(
        "**/R.class",
        "**/R\$*.class",
        "**/BuildConfig.*",
        "**/*\$inlined\$*",
    )
    val kotlinDebug = layout.buildDirectory.dir("tmp/kotlin-classes/debug").get().asFile
    classDirectories.setFrom(
        fileTree(kotlinDebug) {
            exclude(classFilter)
        },
    )
    sourceDirectories.setFrom(files("${layout.projectDirectory}/src/main/java"))
}