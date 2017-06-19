angular.module('harrowApp').filter 'momentDurationFormat', () ->
  (elapsedMilliseconds) ->
    duration = moment.duration(elapsedMilliseconds)
    if duration.hours()
      "#{Math.floor(duration.asHours())}h#{duration.minutes()}m#{duration.seconds()}s"
    else if duration.minutes()
      "#{duration.minutes()}m#{duration.seconds()}s"
    else if duration.seconds()
      "#{duration.seconds()}s"
    else
      "#{duration.milliseconds()}ms"
