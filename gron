#!/usr/bin/env php
<?php
//////////
// Gron //
//////////

// Take valid JSON on stdin, or read it from a file or URL,
// then output it as discrete assignments to make it grep-able.

// Exit codes:
//   0 - Success
//   1 - Failed to decode JSON
//   2 - Argument is not valid file or URL
//   3 - Failed to fetch data from URL

// Tom Hudson - 2012
// https://github.com/TomNomNom/gron 

// Decide on stdin, a local file or URL
if ($argc == 1){
  $buffer = file_get_contents('php://stdin');

} else {
  $source = $argv[1];

  // Check for a readable file or URL 
  if (is_readable($source)){
    $buffer = file_get_contents($source);

  } else if (filter_var($source, FILTER_VALIDATE_URL)){
    $c = curl_init($source);
    curl_setopt($c, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($c, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($c, CURLOPT_TIMEOUT, 20);
    curl_setopt($c, CURLOPT_HTTPHEADER, array(
      'Accept: application/json'
    ));

    $buffer = curl_exec($c);
    $err    = curl_errno($c);
    curl_close($c);

    if ($err != CURLE_OK){
      fputs(STDERR, "Data could not be fetched from [{$source}]\n");
      exit(3);
    }

  } else {
    fputs(STDERR, "[{$source}] is not a valid file or URL.\n");
    exit(2);
  }
}

// Meat
$struct = json_decode($buffer);
$err = json_last_error();

if ($err != JSON_ERROR_NONE){
  // Attempt to read as multiple lines of JSON (sometimes found in streaming APIs etc)
  $lines = explode("\n", trim($buffer));
  for ($i = 0; $i < sizeOf($lines); $i++){

    $line = $lines[$i];
    $struct = json_decode($line);
    $err = json_last_error();

    // No dice; time to die
    if ($err != JSON_ERROR_NONE){
      fputs(STDERR, "Failed to decode JSON\n");
      exit(1);
    }

    printSruct($struct, "json{$i}");
    echo PHP_EOL;
  }
} else {
  // Buffer is all one JSON blob
  printSruct($struct);
}

function printSruct($struct, $prefix = 'json'){
  if (is_object($struct)){
    echo "{$prefix} = {};\n";
  } elseif (is_array($struct)){
    echo "{$prefix} = [];\n";
  } else {
    echo "{$prefix} = ". json_encode($struct) .";\n";

    // No need to iterate if we already have a scalar
    return;
  }

  foreach ($struct as $k => $v){
    $k = json_encode($k);
    printSruct($v, "{$prefix}[{$k}]");
  }
}

exit(0);
