import { useState } from "react"
import { Source } from "../models/source"
import { FilesystemService } from "../services/filesystem-service"

export function useSource(initSources?: Source[]) {
  const [mode, setMode] = useState<"files" | "folder" | "no-select">(
    !initSources || initSources.length === 0 
      ? "no-select" 
      : (initSources.length === 1 
        ? "folder" 
        :"files"
      )
  )
  const [sources, setSources] = useState<Source[]>(initSources ?? [])

  const addFiles = async () => {
    const srcs = await FilesystemService.pickFiles()
    if(!srcs || srcs.length == 0) return

    const newSources = [...sources, ...srcs]
    setSources(newSources)

    if(mode !== "files") setMode("files")
  }
  
  const addFolder = async () => {
    const src = await FilesystemService.pickFolder()
    if (!src) return

    setSources([src])
    if(mode !== "folder") setMode("folder")
  }

  const deleteSource = (index: number) => {
    const newSources = sources.filter((_, idx) => idx !== index)
    setSources(newSources)

    if (newSources.length === 0) setMode("no-select")
  }

  return {mode, sources, addFiles, addFolder, deleteSource}
}