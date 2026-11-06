import java.text.SimpleDateFormat
import java.util.Date

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
}

android {
    namespace = "com.example.grifapp"
    compileSdk {
        version = release(36)
    }

    defaultConfig {
        applicationId = "com.example.grifapp"
        minSdk = 24
        targetSdk = 36
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Настройка фильтров архитектур
        ndk {
            // Мы указываем 4 основные архитектуры, которые генерирует gomobile bind:
            // armeabi-v7a — старые 32-битные Android устройства
            // arm64-v8a   — современные 64-битные смартфоны (основной таргет)
            // x86         — 32-битные эмуляторы
            // x86_64      — 64-битные эмуляторы и хромбуки
            abiFilters.addAll(listOf("armeabi-v7a", "arm64-v8a", "x86", "x86_64"))
        }
    }

    // Логика автоматического переименования APK
    applicationVariants.all {
        val variant = this
        variant.outputs.map { it as com.android.build.gradle.internal.api.ApkVariantOutputImpl }.forEach { output ->
            val date = SimpleDateFormat("yyyy-MM-dd").format(Date())
            val baseName = "grif"
            val version = variant.versionName
            val buildType = variant.name // debug или release

            // Итоговое имя: grif-v1.0.2-2025-12-20-release.apk
            output.outputFileName = "$baseName-v$version-$date-$buildType.apk"
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    kotlinOptions {
        jvmTarget = "11"
    }
}

dependencies {
    //implementation(files("libs/grifapp.aar"))

dependencies {
    implementation(files("libs/mobile.aar"))
    // ... остальные
}

    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.appcompat)
    implementation(libs.material)
    implementation(libs.androidx.activity)
    implementation(libs.androidx.constraintlayout)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
}