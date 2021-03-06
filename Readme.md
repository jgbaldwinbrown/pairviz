# Pairviz

Pairviz is software for converting .pairs files that describe Hi-C data into linear measures of chromosome pairing. There are two working tools now, and more under construction:

## Requirements

Pairviz is written in Python 3 and has the following requirements:

```
matplotlib
numpy
pandas
seaborn
```

## Working tools

### `pairviz.py`

Pairviz converts .pairs files to tab-separated tables that are ready for plotting. The usage is:

```
usage: Count up pairing and chromosome self-interactions in Hi-C .pairs files
       [-h] [-w WINDOW\_SIZE] [-s STEP\_SIZE] [-c] [-i] [-d DISTANCE] [-f]
       [-g GENOME\_LENGTH]
       [input [input ...]]

positional arguments:
  input                 One or more .pairs files to use as input (default =
                        stdin).

optional arguments:
  -h, --help            show this help message and exit
  -w WINDOW\_SIZE, --window\_size WINDOW\_SIZE
                        The size of the sliding window to calculate (default =
                        100kb).
  -s STEP\_SIZE, --step\_size STEP\_SIZE
                        The distance to slide the window each step (default =
                        10kb).
  -c, --chromosome      Ignore sliding window analysis and perform a
                        chromosome-wide count (default = False).
  -i, --standard\_input  Take standard input and other input files (default =
                        False)
  -d DISTANCE, --distance DISTANCE
                        Distance away that two reads can be before they are
                        ignored (default = 5Mb)
  -f, --no\_fpkm         Calculate fpkm along with counts (default = False)
  -g GENOME\_LENGTH, --genome\_length GENOME\_LENGTH
                        Genome size for purpose of fpkm calculations (no
                        default; required if fpkm=True)
```

### `pairviz_plot.py`

Pairviz\_plot converts the tabular output of Pairviz into plots. Its usage is as follows:

```
usage: Visualize Hi-C pairing rates as a 2-d line plot. [-h] [-o OUTPUT]
                                                        [-t TITLE] [-p] [-f]
                                                        [-s] [-c CHROMSPACE]
                                                        [-l] [-i] [-n NAMES]
                                                        [-x X_AXIS_NAME]
                                                        [-y Y_AXIS_NAME]
                                                        [input [input ...]]

positional arguments:
  input                 Input file(s) generated by pairviz (default = stdin).

optional arguments:
  -h, --help            show this help message and exit
  -o OUTPUT, --output OUTPUT
                        output path (default = out.pdf).
  -t TITLE, --title TITLE
                        Title of plot (default = "Pairing Rate").
  -p, --proportion      If included, plot as a proportion of total reads in
                        the region, rather than absolute (default = False).
  -f, --no\_fpkm         If included, plot FPKM rather than read counts
                        (default = False, can be combined with --proportion).
  -s, --self            Also plot self-interactions (default = False).
  -c CHROMSPACE, --chromspace CHROMSPACE
                        bp of space to put between chromosomes in plot
                        (default = 5000000).
  -l, --log             Log-scale the y-axis (default = False).
  -i, --stdin           take input from stdin along with other inputs.
  -n NAMES, --names NAMES
                        Names to use for plotting of input files (default =
                        alphabetical letters).
  -x X\_AXIS\_NAME, --x\_axis\_name X\_AXIS\_NAME
                        X axis name.
  -y Y\_AXIS\_NAME, --y\_axis\_name Y\_AXIS\_NAME
                        Y axis name.
```
