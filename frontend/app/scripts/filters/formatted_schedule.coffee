angular.module('harrowApp').filter 'formattedSchedule', (moment, later) ->
  (schedule, format) ->
    now = null
    if schedule.subject.cronspec
      later.date.localTime();
      now = moment(later.schedule(later.parse.cron(schedule.subject.cronspec)).next(1))
    if schedule.subject.timespec
      now = moment()
      parts = schedule.subject.timespec.split(' + ')
      parts.forEach (sec) ->
        part = sec.split(' ')
        now.add(part[0], part[1])
      now
    return unless now
    if format == 'fromNow'
      now.fromNow(true)
    else if format == 'calendar'
      now.calendar()
    else
      now.format(format)
