package com.tuapp.filehandler

import android.net.Uri
import android.os.Build
import android.content.Intent
import android.os.FileUtils
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import org.apache.commons.io.FilenameUtils
import expo.modules.kotlin.exception.Exceptions
import java.io.File
import java.io.FileNotFoundException
import java.io.FileOutputStream
import java.io.IOException
import java.util.UUID // Importación necesaria para el ID único

class FileHandlerModule : Module() {

  // Función privada que asegura la existencia del directorio
  private fun ensureDirExists(dir: File): File {
    if (!(dir.isDirectory || dir.mkdirs())) {
      throw IOException("Couldn't create directory '$dir'")
    }
    return dir
  }
  
  // Genera la ruta de salida de forma segura
  private fun generateOutputPath(internalDirectory: File, dirName: String, extension: String): String {
    val directory = File(internalDirectory, dirName) // Uso de constructor más limpio
    ensureDirExists(directory)
    
    val filename = UUID.randomUUID().toString()
    // Reemplazo del operador ternario de Java por if de Kotlin
    val formattedExtension = if (extension.startsWith(".")) extension else ".$extension"
    
    return File(directory, filename + formattedExtension).absolutePath
  }

  override fun definition() = ModuleDefinition {
    Name("FileHandler")

    AsyncFunction("copyToCache") { documentUri: String, name: String ->
      val context = appContext.reactContext ?: throw Exceptions.AppContextLost()
      val uri = Uri.parse(documentUri)

      try {
        val takeFlags = Intent.FLAG_GRANT_READ_URI_PERMISSION or 
                       Intent.FLAG_GRANT_WRITE_URI_PERMISSION
        context.contentResolver.takePersistableUriPermission(uri, takeFlags)
      } catch (e: SecurityException) {
        // Si ya tiene permisos o no se pueden tomar, continuar
        android.util.Log.w("FileHandler", "Could not take persistable permissions: ${e.message}")
      }

      val outputFilePath = generateOutputPath(
        context.cacheDir,
        "ToSend",
        FilenameUtils.getExtension(name)
      )
      val outputFile = File(outputFilePath)

      // El método openInputStream requiere un objeto Uri
      context.contentResolver.openInputStream(uri).use { inputStream ->
        if (inputStream == null) throw FileNotFoundException("InputStream for $documentUri was null.")
        
        FileOutputStream(outputFile).use { outputStream ->
          if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            FileUtils.copy(inputStream, outputStream)
          } else {
            inputStream.copyTo(outputStream)
          }
        }
      }
      
      // Retornamos el path como string o el URI según lo que necesite tu JS
      return@AsyncFunction Uri.fromFile(outputFile).toString()
    }
  }
}