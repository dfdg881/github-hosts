import { Bindings } from "./types"
import { fetchLatestHostsData, storeData } from "./services/hosts"

export async function handleSchedule(
  event: ScheduledEvent,
  env: Bindings
): Promise<void> {
  console.log("Running scheduled task...")

  try {
    const newEntries = await fetchLatestHostsData()
    await storeData(env, newEntries)

    console.log("Scheduled task completed successfully")
  } catch (error) {
    console.error("Error in scheduled task:", error)
  }
}
