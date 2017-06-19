Glossary
========

Operation
---------

  An invocation of a Job_. Used for past and present Operations. Schedules_
  create Operations_.

Job
---

  An invocation of a Job_. Used for past and present Operations. Schedules_
  create Operations

Schedule
--------

  A Schedule can either be `One Time`_ or Recurring_. Schedules are the
  mechanism by which an Operation_ is started.

Schedule: Recurring
-------------------

  A Recurring Schedule is used to run something periodically. The recurring
  schedule refers to a Job_ and includes a Cronspec_. The cronspec is evaluated
  in the UTC timezone.

Schedule: One-Time
------------------

  One-time schedules are schedules such as ``now`` or ``now + 5 minutes``, they
  are in Atspec_ format, when clicking "Run Now" in the Harrow interface, a
  one-time schedule is created for "now", meaning that we approach it as soon
  as possible.

Cronspec
--------

  For more information about the `Cronspec please refer to Wikipedia`_.

Atspec
------

  For more information about the `Atspec`

.. _Cronspec please refer to Wikipedia: https://en.wikipedia.org/wiki/Cron#Overview
.. _Job: #job
.. _Jobs: #job
.. _One Time: #schedule-one-time
.. _Operations: #operation
.. _Recurring: #schedule-recurring
.. _Schedule: #schedule
.. _Schedules: #schedule
