package schedule

func (s *Scheduler) PrepareJobs() {
	s.Cron.AddFunc(RUN_EVERY_HOUR_AT_MIN_5, func() {
		// clean up stale free games every hour at minute 5
		s.TriggerCleanup()
	})

	s.Cron.AddFunc(RUN_EVERY_HOUR, func() {
		// get free games update from the Epic Games Store API
		s.TriggerCrawlFreeGamesData()
	})

	s.Cron.AddFunc(RUN_EVERY_HOUR, func() {
		// send free games update to the WhatsApp
		s.TriggerSendFreeGamesUpdate()
	})
}
