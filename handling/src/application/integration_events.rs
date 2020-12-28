use crate::application::pb::{CargoDestinationChanged, NewCargoBooked};
use crate::domain::handling::{Cargo, HandlingEvent, TrackingID};
use crate::domain::Repository;
use crate::Error;
use async_trait::async_trait;
use log::info;
use std::convert::TryInto;

#[async_trait]
pub trait EventService {
    async fn cargo_was_handled(&self, e: HandlingEvent) -> Result<(), Error>;
}

pub trait EventHandler: Clone + Send {
    type Event;
    fn handle(&self, e: Self::Event) -> Result<(), Error>;
}

#[derive(Clone)]
pub struct NewCargoBookedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    cargos: T,
}

impl<T> NewCargoBookedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    pub fn new(cargos: T) -> Self {
        NewCargoBookedEventHandler { cargos }
    }
}

impl<T> EventHandler for NewCargoBookedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    type Event = NewCargoBooked;
    fn handle(&self, e: Self::Event) -> Result<(), Error> {
        let cargo: Cargo = e.try_into()?;
        info!("New cargo booked {}", cargo.tracking_id);
        self.cargos.store(cargo.tracking_id.clone(), &cargo)?;
        Ok(())
    }
}

#[derive(Clone)]
pub struct CargoDestinationChangedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    cargos: T,
}

impl<T> CargoDestinationChangedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    pub fn new(cargos: T) -> Self {
        CargoDestinationChangedEventHandler { cargos }
    }
}

impl<T> EventHandler for CargoDestinationChangedEventHandler<T>
where
    T: Repository<TrackingID, Cargo>,
{
    type Event = CargoDestinationChanged;
    fn handle(&self, e: Self::Event) -> Result<(), Error> {
        info!(
            "Cargo {} destination changed {}",
            e.tracking_id, e.destination
        );
        let mut cargo = self.cargos.find(e.tracking_id)?;
        cargo.destination = e.destination;
        self.cargos.store(cargo.tracking_id.clone(), &cargo)?;
        Ok(())
    }
}