package schedule

func (s *Scheduler) PrepareJobs() {
	s.Cron.AddFunc(RUN_EVERY_DAY_AT_0AM, func() {
		// clean up old data that stale for 2 weeks+
		s.TriggerCleanup()
	})

	s.Cron.AddFunc(RUN_EVERY_3DAYS_AT_0AM, func() {
		// get free games update from the Epic Games Store API
		s.TriggerCrawlFreeGamesData()
	})

	s.Cron.AddFunc(RUN_EVERY_DAY_AT_10AM, func() {
		// send free games update to the WhatsApp
		s.TriggerSendFreeGamesUpdate()
	})
}
