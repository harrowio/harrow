Regular cleanup of the old Git branches
=======================================

While working with Git repositories, it's quite common to get to the point that very many branches are created. 
In such situation, it can be hard or time consuming to identify the branches that have been merged or became unnecessary.

With Harrow, is possible to create a task that periodically cleans up the repository for you:

We wrote a short script that can be made recurrent with a Harrow scheduled task and help in keeping the repository tidy.

In our task we want: 
- to deletes those branches that have already been merged
- to list those branches that have not been merged but have not been changed (no commits) for more than 2 months (this timeframe can be easily changed)

note: in a later release of Harrow it will be possible to send a notification to the branch owners of the inactive branches so that they can take action.

Script is the following:

.. code-block:: bash

  #!/bin/bash    
  #set -x                                                                                 #displays every command executed for troubleshooting
  BRANCHES=$(git branch --list --remote 'origin/*' | grep -v "master" | grep -v -- "->" ) #lists all the branches
  for b in $BRANCHES

  do

  month_of_last_commit=$(git log --pretty='%ci' -n 1 $b | awk -F- '{print $2}')           #gets the month of the last commit
  month=$(date +%m)                                                                       #gets the current month 
  difference_in_months=$(expr $month - $month_of_last_commit)                             #calculates the period of inactivity of the branch   

  if git branch --list --remote 'origin/*' --merged | grep -v "master" | grep -v -- "->" | grep -q $b ; then echo git branch -d $b;
  elif [ $difference_in_months -ge 2 ]; then echo "$b"; taskFail=yes; fi                  #if the branch is merged, delete it, if it's inactive print it to screen

  done

  if [ $taskFail = yes ]; then exit 1;                                                    #make the task fail if there are inactive branches  
  else exit 0; fi

The script will return an exit failure when inactive branches are found.


