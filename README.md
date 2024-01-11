# Twitch Clip Compiler

Twitch Clip Compiler collects clips from your desired Twitch streamer and transforms them into a highlight reel, allowing you to be easily up-to-date with your favorite streamers.
It is available both as a CLI tool and a [web app](https://twitch-clip-compiler.com).

## Prerequisites
The following prerequisites must be satisfied in order to run the CLI tool.
### FFmpeg
#### Windows
  - Download [here](https://www.gyan.dev/ffmpeg/builds/).
  - Extract the folder to a location of your choice.
  - Add FFmpeg to your Path.
    - Copy the path of the `bin` folder located inside the folder you just extracted. For example: `C:\ffmpeg\bin`.
    - Type `Edit the system environment variables` into the start menu search bar.
    - Click on `Environment Variables...`.
    - Under user variables, look for `Path` and click the edit button.
    - Click on `New` and paste the path you just copied earlier.
    - Click `OK` and apply your changes.
  - (Optional) Open the command prompt and enter `ffmpeg -version` to verify that ffmpeg has been installed correctly.
#### Linux
Run the following commands:
```
sudo apt update
sudo apt install ffmpeg
ffmpeg -version
```

### Twitch Client ID and Secret
Follow the steps outlined [here](https://dev.twitch.tv/docs/authentication/register-app/) to get a Twitch Client ID and Secret.

Once you have your Client ID and Secret, add them to your environment with the following names:

| Variable Name         | Value |
| -------------         | -------------       |
| TWITCH_CLIENT_ID      | <YOUR_CLIENT_ID>     |
| TWITCH_CLIENT_SECRET  | <YOUR_CLIENT_SECRET>  |

For a guide on how to add environment variables to your system, kindly refer to the following resources: 
  - [Windows](https://phoenixnap.com/kb/windows-set-environment-variable) 
  - [Linux](https://phoenixnap.com/kb/linux-set-environment-variable)

## Installation
You may download the binary for Windows and Linux from the releases page. 

Alternatively, if you have Go installed, you may also run the following command:

```
go install github.com/jaaanko/twitch-clip-compilation-tool/cmd/clipcompiler@latest
```

## Usage
Enter ```clipcompiler --help``` to display a list of available arguments and options.
```
$ clipcompiler --help

Usage: clipcompiler [options] username start_date end_date
                         
Arguments

        username      :   Unique twitch username of the user you wish to watch clips from. [required]
        start_date    :   Start date in YY-MM-DD format (example: 2023-04-26). [required]
        end_date      :   End date in YY-MM-DD format (example: 2023-04-26). [required]
Options

        --max         :   Maximum number of clips to fetch, no more than 20. Default is 10.
        --output-dir  :   Name of the directory where the final .mp4 file and any temporary files will be placed. 
                          A default folder named "out" will be created in the current directory if not specified.
        --output-file :   Name of the final .mp4 file. Default is "compilation.mp4".
        --help        :   Displays this message and exits the program.
```

### Basic Examples
Fetch a default of 10 clips from `streamer1` within `2023-12-14` to `2023-12-15` :

```
clipcompiler streamer1 2023-12-14 2023-12-15
```

Fetch 5 clips from `streamer1` within `2023-12-14` to `2023-12-15`. Additionally, name the final `mp4` file as `streamer1_clips.mp4` :

```
clipcompiler --max=5 --output-file=streamer1_clips.mp4 streamer1 2023-12-14 2023-12-15
```

Note that if a streamer has less clips available than what was specified in the `max` option, the program will just fetch as much clips as it can.

## Contributing
If you have any issues or suggestions for new features, please feel free to [create a new issue](https://github.com/jaaanko/twitch-clip-compilation-tool/issues/new) or directly contribute. Any feedback on this project is highly appreciated!
