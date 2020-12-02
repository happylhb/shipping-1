use handling::application::grpc_server::{HandlingServiceImpl, HandlingServiceServer};
use handling::application::service::{EventHandlerImpl, ServiceImpl};
use handling::domain::handling::HandlingEventFactoryImpl;
use handling::domain::{location, Result};
use handling::infrastructure::inmem_repository::InmemRepository;
use structopt::StructOpt;
use tonic::transport::Server;

use handling::domain::{handling::Cargo, handling::Voyage, Repository};

/// Handling service
#[derive(StructOpt, Debug)]
struct Opt {
    /// Binding address
    #[structopt(
        long,
        env = "ADDR",
        default_value = "[::1]:50051",
        hide_env_values = true
    )]
    addr: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    let opt = Opt::from_args();

    // prerare dependencies
    let cargos = InmemRepository::new();
    cargos.store("001".to_string(), &Cargo {}).unwrap();
    let voyages = InmemRepository::new();
    voyages.store("v001".to_string(), &Voyage {}).unwrap();
    let locations = InmemRepository::new();
    location::store_sample_locations(&locations)?;
    let handling_events = InmemRepository::new();
    let unimp_event_handler = EventHandlerImpl {};
    let event_factory = HandlingEventFactoryImpl::new(cargos, voyages, locations);

    let srv = ServiceImpl::new_service(handling_events, event_factory, unimp_event_handler);

    let addr = opt.addr.parse()?;

    let gservice = HandlingServiceImpl::new(srv);

    Server::builder()
        .add_service(HandlingServiceServer::new(gservice))
        .serve(addr)
        .await?;

    Ok(())
}
