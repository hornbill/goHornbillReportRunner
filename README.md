# Hornbill Report Runner

This utility provides a quick and easy method of running reports on your Hornbill instance, and retrieving the report output files for storage locally.

The tool will run pre-built reports on your Hornbill instance, wait for the reports to complete, then retrieve all files that were created as part of the report run (as defined within the report itself, for example PDF, CSV, XLS etc).

## Quick Links
- [Installation](#installation)
- [Configuration](#configuration)
- [Execute](#execute)

## Installation

#### Windows
* Download the ZIP archive containing the executables, configuration file and license;
* Extract the ZIP archive into a folder you would like the application to run from e.g. 'C:\hornbill_reporting\'.

## Configuration

Example JSON File:

```
{
    "APIKey": "yourapikey",
    "InstanceID": "yourinstanceid",
    "Reports":[
        {
            "ReportID":6,
            "ReportName":"Change Requests",
            "DeleteReportInstance": true,
            "ReportFolder":"//some/network/drive"
        },
        {
            "ReportID":5,
            "ReportName":"Scheduled Changes",
            "DeleteReportInstance": true,
            "ReportFolder":"//some/network/drive"
        }
    ]
}
```

- "APIKey" - A Valid API Assigned to a user with enough rights to run and retrieve the report (case sensitive)
- "InstanceId" - Instance ID (case sensitive)
- "Reports" - An array containing objects defining the reports to be run and retrieved:
        - "ReportID" - The integer ID of the report to be run
        - "ReportName" - The name of the report to be run
        - "DeleteReportInstance" - Boolean true or false, to define if the report run instance should be removed after the report run has completed
        - "ReportFolder" - The network or local drive location where you with to save the output report files

## Execute

Command Line Parameter:
- file
This should point to your json configration file and by default looks for a file in the current working directory called conf.json. If this is present you don't need to have the parameter.

'goHornbillReportRunner_x64.exe -file=conf.json'

## Preparing to run the tool

- Open '''conf.json''' and add in the necessary configration;
- Open Command Line Prompt as Administrator;
- Change Directory to the folder with goHornbillGetReport_* executables 'C:\hornbill_reporting\';
- Run the command: 
        - On 32 bit Windows PCs: goHornbillReportRunner_x86.exe
        - On 64 bit Windows PCs: goHornbillReportRunner_x64.exe
- Follow all on-screen prompts, taking careful note of all prompts and messages provided.
