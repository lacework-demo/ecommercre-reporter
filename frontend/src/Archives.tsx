import React, {useEffect} from "react";
import {Link, useLocation} from "react-router-dom";
import {useRecoilCallback, useRecoilRefresher_UNSTABLE, useRecoilValue, useSetRecoilState} from "recoil";
import "./Archives.css";
import {getArchive, getArchives, neverMountedArchives} from "./data";
import {BarChart, OrderListing} from "./Orders";

function useQuery() {
  const {search} = useLocation();
  return React.useMemo(() => new URLSearchParams(search), [search]);
}

export function ArchiveReport() {
  const archiveUrl = useQuery().get("archiveUrl")
  const archiveData = useRecoilValue(getArchive(archiveUrl))


  return <div className="archive-body">
    <span>Archive Report ID: {archiveUrl?.split("/archive/")[1]}</span>
    <br />
    <BarChart chartData={archiveData?.chartData} />
    <OrderListing orders={archiveData?.orders} />
  </div>

}

export function ArchivesBody() {
  const archiveCleanup = useRecoilRefresher_UNSTABLE(getArchives)
  const setMountedState = useSetRecoilState(neverMountedArchives)
  const neverMountedArchivesState = useRecoilCallback(({snapshot}) => async () => {
    const nm = await snapshot.getPromise(neverMountedArchives);
    return nm
  }, [])

  useEffect(() => {
    async function setMounted() {
      if (await neverMountedArchivesState()) {
        setMountedState(false)
      } else {
        archiveCleanup()
      }

    }
    setMounted()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return <div><Archives /></div>
}

export function Archives() {
  const data = useRecoilValue(getArchives)

  if (data.length > 0) {
    return <div className="archive-body">
      <h2 style={{textAlign: "center"}}>Archives</h2>
      <table>
        <thead>
          <tr>
            <th>url</th>
          </tr>
        </thead>
        <tbody>
          {data.map(d =>
            <tr key={d.url}>
              <td>
                <Link className="link" to={`/archiveReport?archiveUrl=/archive/${d.url}`}>
                  {d.name}
                </Link>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  } else {
    return <div className="archive-body">
      <h2 style={{textAlign: "center"}}>Archives</h2>
      <span>No data</span>
    </div>
  }
}
