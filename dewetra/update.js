const loadFile = (file) => require(`${__dirname}/../data/${file}.json`);
const stationsById = loadFile("stations");
const stations = Object.values(stationsById);
const eustations = loadFile("euro-stations");
const stationsByCoords = {};

for (const st of eustations) {
  if (!(st.ID in stationsById)) {
    console.log("solo in eustations:", st.ID);
  } else {
    stationsById[st.ID].done = true;
  }
}

for (const st of stations) {
  if (!st.done) {
    console.log("solo in stations:", st.ID);
  }
}

process.exit();

const mkKey = (lat, lon, name, prec) =>
  `${Math.round(lat * prec) / prec}:${Math.floor(lon * prec) / prec}`;

const mkKey2 = (lat, lon, name, prec) =>
  `${Math.round(lat * prec) / prec}:${Math.round(lon * prec) / prec}`;

const mkKey3 = (lat, lon, name, prec) =>
  `${Math.floor(lat * prec) / prec}:${Math.round(lon * prec) / prec}`;

const mkKey4 = (lat, lon, name, prec) =>
  `${Math.floor(lat * prec) / prec}:${Math.floor(lon * prec) / prec}`;

for (const st of stations) {
  const key = mkKey(st.Latitude, st.Longitude, st.Neighborhood, 100000);
  //console.log("orig:", key);
  stationsByCoords[key] = st;
}

let dupl = 0;
const updateByClass = (name) => {
  const anemometro = loadFile(name);
  for (const anem of anemometro) {
    let key;
    let st;

    for (let prec = 1000000; prec >= 100; prec /= 10) {
      key = mkKey(anem.lat, anem.lon, anem.stationName, prec);
      st = stationsByCoords[key];

      if (!st) {
        console.log("not found:", key, prec);
        key = mkKey2(anem.lat, anem.lon, anem.stationName, prec);
        st = stationsByCoords[key];
      } else {
        break;
      }

      if (!st) {
        console.log("not found:", key, prec);
        key = mkKey3(anem.lat, anem.lon, anem.stationName, prec);
        st = stationsByCoords[key];
      } else {
        break;
      }

      if (!st) {
        console.log("not found:", key, prec);
        key = mkKey4(anem.lat, anem.lon, anem.stationName, prec);
        st = stationsByCoords[key];
      } else {
        break;
      }
    }

    if (!st) {
      //console.log("not found:", anem.lat, anem.lon);
      continue;
    } else {
      //console.log("found:", key);
    }

    if (st.dewetraIDS === undefined) {
      st.dewetraIDS = {};
    }

    if (st.dewetraIDS[name] !== undefined && st.dewetraIDS[name] !== anem.id) {
      console.log(
        `duplicated by coord: ${key}` + st.dewetraIDS[name] + "-" + anem.id
      );
      dupl++;
      continue;
    }

    st.dewetraIDS[name] = anem.id;
  }
};

updateByClass("ANEMOMETRO");
updateByClass("DIREZIONEVENTO");
updateByClass("IGROMETRO");
updateByClass("PLUVIOMETRO");
updateByClass("TERMOMETRO");

for (const st of eustations) {
  const fixst = stationsById[st.ID];
  if (!fixst) {
    console.log("NON trovata:", st.ID);
    continue;
  }

  if (!fixst.dewetraIDS) {
    console.log("senza dewetra:", st.ID);
  }
}

console.log(`NON TROVATE: ${stations.filter((st) => !st.dewetraIDS).length}`);
console.log(`TROVATE: ${stations.filter((st) => !!st.dewetraIDS).length}`);
console.log(`TOTALI: ${stations.length}`);
console.log(`TOTALI EU: ${eustations.length}`);
console.log(`DUPLICATE: ${dupl}`);

//console.log(stations.filter((st) => !st.dewetraIDS));
