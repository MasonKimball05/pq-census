# Stage 7: the report. Rscript report.R writes figures/ and STATS.md.
# See the guide for what it should contain. This file is yours from the start.
for (file in list.files("R", pattern = "\\.R$", full.names = TRUE)) source(file)
