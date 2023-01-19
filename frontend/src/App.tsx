import {Link, Outlet, useRouteError} from 'react-router-dom';
import './App.css';

export function ErrorPage() {
  const error: any = useRouteError();

  return (
    <div className="App">
      <div className="container">
        <div className="header">
          <Menu />
        </div>
        <div className="body">
          <div className="error-body">
            <h2>Ooof, that didn't go well.</h2>
            <br />
            <p>
              <i>{error.statusText || error.message}</i>
            </p>
          </div>
        </div>
      </div>
    </div>
  );

}
function App() {
  return (
    <div className="App">
      <div className="container">
        <div className="header">
          <Menu />
        </div>
        <div className="body">
          <Outlet />
        </div>
      </div>
    </div>
  );
}

function Menu() {
  return (
    <div className="menu">
      <div className="title">
        <ul className="ul-menu">
          <li className="ul-left">
            <Link to={`/`}>eCommerce Reporting Service</Link>
          </li>
          <li className="ul-right">
            <Link to={`/archives`}>Archives</Link>
          </li>
          <li className="ul-right">
            <Link to={`/report`}>Report</Link>
          </li>
        </ul>
      </div>
    </div>
  )
}

export function Loading() {
  return <div className="loading">
    <h1>loading...</h1>
  </div>
}

export default App;

