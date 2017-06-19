Slack Notifier
==============

Harrow can send notifications to let you know about the success or failure of your task with a message to your Slack.
You can easily configure this service as follows:

Generate a Slack Webhook
------------------------

* Login to Slack
* Go to `Slack incoming webhook setup page`_ 
* Select the channel where you want to be notified
* Click “Add Incoming WebHooks Intgration” button

.. figure:: ../quickstart/1createWebhook.png
   :align: center
   :scale: 50 %

Add your Slack webhook to Harrow
--------------------------------

* Login to Harrow
* Select the project  and go to settings

.. figure:: ../quickstart/2settings.png
   :align: center  

* In the “Notifiers” tab, click on add Slack notifier

.. figure:: ../quickstart/3notifiers.png
   :align: center
   :scale: 50 %
   
* Paste the webhook URL that has been generated in Slack to the “webhook URL” field and add the #channel to the field above

*copy the link from Slack*

.. figure:: ../quickstart/4webhookURL.png
   :scale: 50 %
   :align: center

*paste the URL to Harrow* 

.. figure:: ../quickstart/5addURL.png
   :scale: 75 %  
   :align: center 

* Go to your Dashboard and click on the task that you want to be notified about

.. figure:: ../quickstart/6task.png
   :align: center
   :scale: 50 %

* Add a “Notification” by clicking on the + sign

* At the bottom of the page select in which case you want to be notified: “success” or “failure” of your task

.. figure:: ../quickstart/7selectSlack.png
   :align: center
   :scale: 50 %

* Back to your Slack, click on the setting gear and make sure that the "Notifications settings" are set correctly 

.. figure:: ../quickstart/9slackPreferences.png
   :align: center
   :scale: 50 %

* When your task complete, you will receive a notification from Harrow in the selected Slack channel

.. figure:: ../quickstart/result.png
   :align: center

.. _Slack incoming webhook setup page: https://my.slack.com/services/new/incoming-webhook
