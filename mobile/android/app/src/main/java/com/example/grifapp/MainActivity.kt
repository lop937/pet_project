package com.example.grifapp

import android.webkit.JsResult
import android.annotation.SuppressLint
import android.content.ContentValues
import android.os.Build
import android.os.Bundle
import android.os.Environment
import android.provider.MediaStore
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.appcompat.app.AppCompatActivity
import mobile.Mobile
import java.io.File
import java.io.FileOutputStream
import java.net.HttpURLConnection
import java.net.URL

class MainActivity : AppCompatActivity() {
    private val port = 9000

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val dbFile = applicationContext.getDatabasePath("grif.db")
        if (!dbFile.parentFile.exists()) {
            dbFile.parentFile.mkdirs()
        }
        val dbPath = dbFile.absolutePath

        try {
            Mobile.startServer(port.toLong(), dbPath)
        } catch (e: Exception) {
            e.printStackTrace()
        }

        val webView = WebView(this)

        // Нативное окно подтверждения для JS confirm()
        webView.webChromeClient = object : android.webkit.WebChromeClient() {
            override fun onJsConfirm(
                view: WebView?,
                url: String?,
                message: String?,
                result: JsResult?
            ): Boolean {
                val builder = android.app.AlertDialog.Builder(this@MainActivity)
                builder.setTitle("Подтверждение")
                builder.setMessage(message)
                builder.setPositiveButton("Да") { _, _ -> result?.confirm() }
                builder.setNegativeButton("Отмена") { _, _ -> result?.cancel() }
                builder.setOnCancelListener { result?.cancel() }
                builder.show()
                return true
            }
        }

        val settings: WebSettings = webView.settings
        settings.javaScriptEnabled = true
        settings.domStorageEnabled = true
        settings.loadWithOverviewMode = true
        settings.useWideViewPort = true

        webView.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView?, url: String?) {
                super.onPageFinished(view, url)
            }

            override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
                val urlStr = request?.url.toString()

                if (urlStr.contains("/export/")) {
                    Thread {
                        try {
                            val connection = URL(urlStr).openConnection() as HttpURLConnection
                            connection.connectTimeout = 5000
                            connection.readTimeout = 10000
                            connection.requestMethod = "GET"

                            if (connection.responseCode == HttpURLConnection.HTTP_OK) {
                                val inputStream = connection.inputStream
                                val fileName = "grif_export_${System.currentTimeMillis()}.xlsx"
                                var savePath = ""

                                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                                    // Android 10+ (API 29+): Сохраняем в общую папку Downloads через MediaStore
                                    val resolver = contentResolver
                                    val contentValues = ContentValues().apply {
                                        put(MediaStore.MediaColumns.DISPLAY_NAME, fileName)
                                        put(MediaStore.MediaColumns.MIME_TYPE, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
                                        put(MediaStore.MediaColumns.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
                                    }
                                    val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, contentValues)
                                    if (uri != null) {
                                        resolver.openOutputStream(uri)?.use { outputStream ->
                                            inputStream.copyTo(outputStream)
                                        }
                                        savePath = "Загрузки/$fileName"
                                    }
                                } else {
                                    // Android 9 и ниже: Прямая запись в общую папку Downloads
                                    val downloadsDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
                                    val file = File(downloadsDir, fileName)
                                    FileOutputStream(file).use { outputStream ->
                                        inputStream.copyTo(outputStream)
                                    }
                                    savePath = file.absolutePath
                                }

                                inputStream.close()

                                runOnUiThread {
                                    android.widget.Toast.makeText(this@MainActivity, "Файл сохранён: $savePath", android.widget.Toast.LENGTH_LONG).show()
                                }
                            } else {
                                runOnUiThread {
                                    android.widget.Toast.makeText(this@MainActivity, "Ошибка сервера: ${connection.responseCode}", android.widget.Toast.LENGTH_LONG).show()
                                }
                            }
                        } catch (e: Exception) {
                            runOnUiThread {
                                android.widget.Toast.makeText(this@MainActivity, "Ошибка скачивания: ${e.message}", android.widget.Toast.LENGTH_LONG).show()
                            }
                        }
                    }.start()
                    return true
                }

                // Перехват POST-запроса на выгрузку сметы
                if (urlStr.contains("/export/smeta-selected")) {
                    // Тут мы не можем прочитать POST-данные, поэтому
                    // мы перехватим саму отправку формы через JS-интерфейс (см. шаг 3)
                }

                return false
            }
        }

        setContentView(webView)
        webView.loadUrl("http://127.0.0.1:$port/work")
    }
}